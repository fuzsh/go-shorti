#Go-Shorti
Go-Shorti allow you to run Your Own URL Shortener, on your server.

### QUICK START
**Features:**

- build link
    - unique link per user if duplicated link
    - can use api with `similar_to` option to create short link based on what you want

`{
    "url": "https://google.com",
    "similar_to": "google"
}`

- analytics
    - daily, monthly, weekly
    - uniq, overall
    

**Technologies:**
- golang
- redis
    - redirection
    - save user verify code
- postgres
    - store all hits
    - user info
- rabbitmq
    - send email in a queue (microservice style)
    
**Start:**
- manually
    - create postgres database
    - change the conf file to run locally
    - run the email microservice in port 8084
    - run code
    
- docker 
    - dockerfile to build image (need prod.yml in config folder)
    - change microservice conf then build
    
**TODO**
- [ ] fully dockerize (soon)
- [ ] null response in analytics if there is no hit
- [ ] implement store in the better way, maybe it is the best :))

**Implementable approaches to improvement:**
- [ ] have cron job to delete all the last day hits and store them in single row in database
    