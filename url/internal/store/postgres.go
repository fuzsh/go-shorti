package store

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"strings"
	"url/internal/analytics"
	"url/internal/auth"
	"url/internal/track"
	"url/pkg/log"
)

type PostgresConfig struct {
	Logger   log.Logger
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// Store implements the Store interface.
type PostgresStore struct {
	DB     *sqlx.DB
	logger log.Logger
}

// NewPostgresStore creates a new store storage for given database connection and logger.
func NewPostgresStore(conf PostgresConfig) (*PostgresStore, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", conf.Host, conf.Port, conf.User, conf.Password, conf.DBName)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	//defer db.Close()
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresStore{
		DB:     sqlx.NewDb(db, "postgres"),
		logger: conf.Logger,
	}, nil
}

// NewTx implements the Store interface.
func (store *PostgresStore) NewTx() *sqlx.Tx {
	tx, err := store.DB.Beginx()
	if err != nil {
		store.logger.Errorf("error creating new transaction: %s", err)
	}
	return tx
}

// Commit implements the Store interface.
func (store *PostgresStore) Commit(tx *sqlx.Tx) {
	if err := tx.Commit(); err != nil {
		store.logger.Infof("error committing transaction: %s", err)
	}
}

// Rollback implements the Store interface.
func (store *PostgresStore) Rollback(tx *sqlx.Tx) {
	if err := tx.Rollback(); err != nil {
		store.logger.Infof("error rolling back transaction: %s", err)
	}
}

// SaveHits implements the Store interface.
func (store *PostgresStore) SaveHits(hits []track.Hit) error {
	const hitParams = 19
	args := make([]interface{}, 0, len(hits)*hitParams)
	var query strings.Builder
	query.WriteString(`INSERT INTO "hit" (tenant_id, fingerprint, session, path, url, language, user_agent, referrer, os, os_version, browser, browser_version, country_code, desktop, mobile, screen_width, screen_height, screen_class, time) VALUES `)

	for i, hit := range hits {
		args = append(args, hit.TenantID)
		args = append(args, hit.Fingerprint)
		args = append(args, hit.Session)
		args = append(args, hit.Path)
		args = append(args, hit.URL)
		args = append(args, hit.Language)
		args = append(args, hit.UserAgent)
		args = append(args, hit.Referrer)
		args = append(args, hit.OS)
		args = append(args, hit.OSVersion)
		args = append(args, hit.Browser)
		args = append(args, hit.BrowserVersion)
		args = append(args, hit.CountryCode)
		args = append(args, hit.Desktop)
		args = append(args, hit.Mobile)
		args = append(args, hit.ScreenWidth)
		args = append(args, hit.ScreenHeight)
		args = append(args, hit.ScreenClass)
		args = append(args, hit.Time)
		index := i * hitParams
		query.WriteString(fmt.Sprintf(`($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d),`,
			index+1, index+2, index+3, index+4, index+5, index+6, index+7, index+8, index+9, index+10, index+11, index+12, index+13, index+14, index+15, index+16, index+17, index+18, index+19))
	}

	queryStr := query.String()
	_, err := store.DB.Exec(queryStr[:len(queryStr)-1], args...)

	if err != nil {
		return err
	}

	return nil
}

func (store *PostgresStore) CreateUser(user auth.User) error {
	query := `INSERT INTO "users" (username, password, is_verified) VALUES(:username, :password, :is_verified)`
	_, err := store.DB.NamedExec(query, user)
	return err
}

func (store *PostgresStore) FindOneByEmail(email string) (auth.User, error) {
	query := `SELECT * FROM users WHERE username = $1`
	var user auth.User
	if err := store.DB.Get(&user, query, email); err != nil {
		return auth.User{}, err
	}
	return user, nil
}

func (store *PostgresStore) VerifyEmail(tx *sqlx.Tx, email string) error {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}
	_, err := tx.Exec(`UPDATE users SET "is_verified" = TRUE WHERE username = $1`, email)
	return err
}

func (store *PostgresStore) CreateLink(tx *sqlx.Tx, url, path string) (int, error) {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}
	var linkID int
	err := tx.Get(&linkID,`INSERT INTO links (url, shortner_path) values($1, $2) RETURNING link_id `, url, path)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	return linkID, nil
}

func (store *PostgresStore) CreateUserLinkRelation(tx *sqlx.Tx, userID int, linkID int) error {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}
	_, err := tx.Exec(`INSERT INTO user_links (user_id, link_id) values($1, $2)`, userID, linkID)
	return err
}

func (store *PostgresStore) GetAnalytics(tx *sqlx.Tx, conf analytics.Config, userID int) (interface{}, error) {
	var query string
	var time string
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}
	if conf.Date == "daily" {
		time = " "
	} else if conf.Date == "yesterday" {
		time = "- interval '1 day' "
	} else if conf.Date == "weekly" {
		time = "- interval '7 days' "
	} else {
		time = "- interval '1 month' "
	}
	if conf.Unique {
		query += `WITH hit_with_time
		AS
		(
			SELECT DISTINCT "fingerprint", path, desktop, mobile, browser from hit where hit.time > CURRENT_DATE ` + time +
			`)
 		SELECT count(distinct "fingerprint") as visitors,`
	} else {
		query += `WITH hit_with_time
		AS
		(
			SELECT fingerprint, path, desktop, mobile, browser from hit where hit.time > CURRENT_DATE ` + time +
			`)
		SELECT count(fingerprint) as visitors,`
	}
	if conf.Mode == "all" {
		query += ` (select count(browser) from hit_with_time where browser = 'Chrome' and path = h.path) As browser_chrome,
				(select count(browser) from hit_with_time where browser = 'Firefox' and path = h.path) As browser_firefox,
				(select count(browser) from hit_with_time where browser <> 'Firefox' and browser <> 'Chrome' and path = h.path) As browser_others,
				(select count(desktop) from hit_with_time where desktop IS TRUE and path = h.path) As platform_desktop,
				(select count(mobile) from hit_with_time where mobile IS TRUE and path = h.path) As platform_mobile,`
	} else if conf.Mode == "platform" {
		query += ` (select count(desktop) from hit_with_time where desktop IS TRUE and path = h.path) As platform_desktop,
				(select count(mobile) from hit_with_time where mobile IS TRUE and path = h.path) As platform_mobile,`
	} else {
		query += ` (select count(browser) from hit_with_time where browser = 'Chrome' and path = h.path) As browser_chrome,
				(select count(browser) from hit_with_time where browser = 'Firefox' and path = h.path) As browser_firefox,
				(select count(browser) from hit_with_time where browser <> 'Firefox' and browser <> 'Chrome' and path = h.path) As browser_others,`
	}
	query += `h.path from users
		inner join user_links ul on ul.user_id = users.user_id
		inner join links l on l.link_id = ul.link_id
		inner join hit_with_time h on h.path = l.shortner_path
		where users.user_id = $1
		group by h.path`

	if conf.Mode == "all" {
		var stats []analytics.Stats
		if err := tx.Select(&stats, query, userID); err != nil {
			return nil, err
		}
		return stats, nil
	}

	if conf.Mode == "platform" {
		var stats []analytics.StatsPlatformMode
		if err := tx.Select(&stats, query, userID); err != nil {
			return nil, err
		}
		return stats, nil
	}

	var stats []analytics.StatsBrowserMode
	if err := tx.Select(&stats, query, userID); err != nil {
		return nil, err
	}
	return stats, nil
}
