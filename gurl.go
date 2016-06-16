package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "net/url"
    "github.com/garyburd/redigo/redis"
)

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

func handler(w http.ResponseWriter, r *http.Request, c redis.Conn) {
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
                fmt.Fprintf(w, "<html><head><title>created</title></head><body>created %d <a href=\"/%s\">http://localhost:8080/%s</a> for %s</body></html>", n, key, key, targetUrl)
            }
    	} else {
            fmt.Fprintf(w, "error creating shortened URL")
    	}
    }
}

func main() {
    c, err := redis.Dial("tcp", ":6379")
    if err != nil {
        // handle error
    }
    defer c.Close()
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        handler(w, r, c)
    })
    http.ListenAndServe(":8080", nil)
}
