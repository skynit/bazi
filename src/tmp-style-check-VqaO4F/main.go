package main
import (
 "context"; "fmt"; "log"; "strings"
 "bazi/internal/service/bazi"; "bazi/internal/service/interpretation"; "bazi/internal/service/localrag"; "bazi/internal/store"
 "gorm.io/driver/sqlite"; "gorm.io/gorm"
)
func main(){ db,err:=gorm.Open(sqlite.Open("../data/bazi.db"),&gorm.Config{}); if err!=nil{log.Fatal(err)}; svc:=&interpretation.Service{Charts:store.NewDBChartStore(db),Bazi:&bazi.BaziService{},Retriever:localrag.NewRetriever(localrag.Config{Enabled:true,IndexPath:"../data/bazi_fts.db",MinScore:0.35,TopK:8}),MinScore:0.35,TopK:8}; resp,err:=svc.InterpretBazi(context.Background(),interpretation.Request{ChartID:82,UserID:2,Focus:"overview"}); if err!=nil{log.Fatal(err)}; fmt.Println(resp.Sections[0].Title); fmt.Println(first(resp.Sections[0].Content, 900)); fmt.Println("--- raw newlines:", strings.Count(resp.Sections[0].Content,"\n")) }
func first(s string,n int)string{r:=[]rune(s); if len(r)<=n{return s}; return string(r[:n])+"..."}
