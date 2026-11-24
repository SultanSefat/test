package main

import (
  "context"
  "encoding/json"
  "log"
  "net/http"
  "os"
  "strings"
  "time"

  "github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var db *redis.Client

type ChatRequest struct { SessionID string `json:"sessionId"`; Message string `json:"message"` }
type ChatResponse struct { Answer string `json:"answer"`; Sources []string `json:"sources"` }

func jsonOut(w http.ResponseWriter, v any) { w.Header().Set("Content-Type","application/json"); json.NewEncoder(w).Encode(v) }
func main() {
  addr := os.Getenv("REDIS_URL")
  if addr == "" { addr = "localhost:6379" }
  opt, err := redis.ParseURL("redis://" + addr)
  if strings.Contains(addr, "://") { opt, err = redis.ParseURL(addr) }
  if err != nil { opt = &redis.Options{Addr: addr} }
  db = redis.NewClient(opt)
  mux := http.NewServeMux()
  mux.HandleFunc("/api/health", func(w http.ResponseWriter,r *http.Request){ jsonOut(w,map[string]string{"status":"ok"}) })
  mux.HandleFunc("/api/knowledge", func(w http.ResponseWriter,r *http.Request){
    if r.Method == http.MethodPost { var v map[string]string; if json.NewDecoder(r.Body).Decode(&v)==nil { db.Set(ctx,"knowledge",v["content"],0) }; jsonOut(w,map[string]bool{"ok":true}); return }
    v,_ := db.Get(ctx,"knowledge").Result(); jsonOut(w,map[string]string{"content":v})
  })
  mux.HandleFunc("/api/chat", func(w http.ResponseWriter,r *http.Request){
    var q ChatRequest
    if json.NewDecoder(r.Body).Decode(&q)!=nil || strings.TrimSpace(q.Message)=="" { http.Error(w,"invalid request",400); return }
    knowledge,_ := db.Get(ctx,"knowledge").Result()
    answer := "I don't have verified information about that in the university materials yet. Please contact the admissions office for confirmation."
    msg := strings.ToLower(q.Message)
    if knowledge!="" {
      terms := strings.Fields(msg); best := ""; score := 0
      for _, p := range strings.Split(knowledge,"\n\n") { s:=0; for _,t:=range terms { if len(t)>3 && strings.Contains(strings.ToLower(p),t) { s++ } }; if s>score { score=s; best=p } }
      if score>0 { answer=best }
    }
    if q.SessionID=="" { q.SessionID="default" }
    b,_:=json.Marshal(map[string]string{"q":q.Message,"a":answer,"at":time.Now().Format(time.RFC3339)})
    db.RPush(ctx,"history:"+q.SessionID,string(b))
    jsonOut(w,ChatResponse{Answer:answer,Sources:[]string{}})
  })
  mux.HandleFunc("/api/history",func(w http.ResponseWriter,r *http.Request){ id:=r.URL.Query().Get("sessionId"); if id=="" {id="default"}; v,_:=db.LRange(ctx,"history:"+id,0,-1).Result(); jsonOut(w,v) })
  port:=os.Getenv("PORT"); if port=="" {port="8080"}
  log.Fatal(http.ListenAndServe(":"+port,mux))
}
