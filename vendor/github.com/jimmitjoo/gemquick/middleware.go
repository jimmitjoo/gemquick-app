package gemquick

import "net/http"

func (g *Gemquick) SessionLoad(next http.Handler) http.Handler {
	g.InfoLog.Println("SessionLoad called")
	return g.Session.LoadAndSave(next)
}
