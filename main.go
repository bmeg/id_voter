package main
import (
  "log"
  "flag"
  "net/http"
  "math/rand"
  "net/url"
  "text/template"
  "context"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
)

func randomSet(c, n int) []int {
  o := rand.Perm(n)
  return o[0:c]
}

type Page struct {
  Email string
  Term  string
  Suggestions map[string]string
}

var message = `
<html>
<form action="/" method="get" id="form1">
<div>
Email<input name="email" type="text" value={{.Email}}></input>
<div>
Term: {{.Term}}
</div>
<div>
<input type="hidden" name="term" value={{.Term}}/>
</form>
<div>
{{ range $key, $value := .Suggestions }}
  <div>
   <button name="vote" type="submit" form="form1" value="{{$key}}">{{$value}}</button>
  </div>
{{ end }}
</div>
</html>
`

type VoteServer struct {
  client *mongo.Client
}

func (vs VoteServer) serve(w http.ResponseWriter, r *http.Request) {
  email := ""
  m, _ := url.ParseQuery(r.URL.RawQuery)
  if x, ok := m["email"]; ok {
    email=x[0]
  }

  if term, ok := m["term"]; ok {
    if vote, ok := m["vote"]; ok {
      log.Printf("%s %s", term, vote)
    }
  }

  coll := vs.client.Database("idvote").Collection("terms")

  ctx := context.Background()

  agg := bson.M{ "$sample" : bson.M{ "size" : 1 } }
  cur, err := coll.Aggregate(ctx, []bson.M{agg})
  if err != nil {
    log.Printf("Error: %s", err)
    return
  }

  elem := map[string]interface{}{}
  for cur.Next(ctx) {
    if err := cur.Decode(&elem); err != nil {
  		log.Fatal(err)
  	}
    log.Printf("%s", elem)
  }
  term := ""
  if t, ok := elem["term"]; ok {
    if ts, ok := t.(string); ok {
      term = ts
    }
  }
  sugg := map[string]string{}
  if sk, ok := elem["suggestions"]; ok {
    if skm, ok := sk.(map[string]interface{}); ok {
      for k, v := range skm {
        sugg[k] = v.(string)
      }
    }
  }
  p := Page{Term:term, Suggestions:sugg, Email:email}
  t := template.Must(template.New("form").Parse(message))
  t.Execute(w, p)
}


func main() {
  flag.Parse()

  ctx := context.Background()

  client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
  if err != nil {
  	log.Fatal(err)
  }
  err = client.Connect(ctx)
  if err != nil {
    log.Fatal(err)
  }

  port := flag.Arg(0)

  vs := VoteServer{client}
  http.HandleFunc("/", vs.serve)
  if err := http.ListenAndServe(":" + port, nil); err != nil {
    panic(err)
  }
}
