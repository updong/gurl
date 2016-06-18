package main

import (
    "log"
    "fmt"
    "io/ioutil"
    "net/http"
    "net/url"
    "encoding/json"
    "github.com/garyburd/redigo/redis"
)

type Config struct {
    Port int
    RedisHost string
    RedisPort int
}

const cmap = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func reverse(s string) string {
    runes := []rune(s)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
}

func id2short(n int) string {
  r := ""
  for n > 0 {
    r += string(cmap[n%62])
    n = n/62
  }
  return reverse(r)
}

func handler(w http.ResponseWriter, r *http.Request, c redis.Conn, conf Config) {
    if r.Method == "GET" {
        key := r.URL.Path[1:]
        if key != "" {
    	    url, err := redis.String(c.Do("GET", key))
            if err == nil {
                http.Redirect(w, r, url, http.StatusFound)
            }
        }
        fmt.Fprintf(w, "<html><head><title>welcome</title></head><body><form action=\"/\" method=\"POST\">"+
            "<input type=\"text\" name=\"url\" size=100>"+
            "<input type=\"submit\" value=\"Save\">"+
            "</form></body></html>")
    } else if r.Method == "POST" {
    	n, err := redis.Int(c.Do("INCR", "testcounter1"))
    	if err == nil {
            key := id2short(n)
            body, err := ioutil.ReadAll(r.Body)
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            vals, err := url.ParseQuery(string(body))
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            targetUrl := vals.Get("url")
            if targetUrl != "" {
    	        _, err := c.Do("SET", key, targetUrl)
                if err != nil {
                    fmt.Fprintf(w, "error creating shortened URL")
                }
                fmt.Fprintf(w, "<html><head><title>created</title></head><body>created(id %d) short url <a href=\"/%s\">%s</a> for %s</body></html>", n, key, key, targetUrl)
            }
    	} else {
            fmt.Fprintf(w, "error creating shortened URL")
    	}
    }
}

func getconf(conffile string) Config {
    b, err := ioutil.ReadFile(conffile)
    if err != nil {
        log.Fatal("error reading config")
    }
    conf := Config{}
    err = json.Unmarshal(b, &conf)
    if err != nil {
        log.Fatal("bad config file")
    }
    return conf
}

func main() {
    conf := getconf("config.json")
    c, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", conf.RedisHost, conf.RedisPort))
    if err != nil {
        // handle error
        log.Fatal("Cannot connect to redis exiting")
    }
    defer c.Close()
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        handler(w, r, c, conf)
    })
    http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), nil)
}
