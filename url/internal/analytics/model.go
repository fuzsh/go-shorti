package analytics

type Config struct {
	Unique bool
	Date   string
	Mode   string
}


type Stats struct {
	Path            string `db:"path" json:"path"`
	Visitors        int    `db:"visitors" json:"visitors"`
	PlatformDesktop int    `db:"platform_desktop" json:"platform_desktop"`
	PlatformMobile  int    `db:"platform_mobile" json:"platform_mobile"`
	BrowserChrome   int    `db:"browser_chrome" json:"browser_chrome"`
	BrowserFirefox  int    `db:"browser_firefox" json:"browser_firefox"`
	BrowserOthers   int    `db:"browser_others" json:"browser_others"`
}

type StatsPlatformMode struct {
	Path            string `db:"path" json:"path"`
	Visitors        int    `db:"visitors" json:"visitors"`
	PlatformDesktop int    `db:"platform_desktop" json:"platform_desktop"`
	PlatformMobile  int    `db:"platform_mobile" json:"platform_mobile"`
}

type StatsBrowserMode struct {
	Path            string `db:"path" json:"path"`
	Visitors        int    `db:"visitors" json:"visitors"`
	BrowserChrome   int    `db:"browser_chrome" json:"browser_chrome"`
	BrowserFirefox  int    `db:"browser_firefox" json:"browser_firefox"`
	BrowserOthers   int    `db:"browser_others" json:"browser_others"`
}
