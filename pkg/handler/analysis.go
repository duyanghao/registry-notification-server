package handler

import (
	"fmt"
	"github.com/duyanghao/registry-notification-server/config"
	"github.com/duyanghao/registry-notification-server/models"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"os"
	"strings"
	"sync"
)

var analysis_lock *sync.Mutex = new(sync.Mutex)

func ProcessAnalysis(w http.ResponseWriter, r *http.Request, c *config.Config) {
	uri := r.RequestURI
	if uri == "/analysis/" {
		http.ServeFile(w, r, "./pages/analysis/home.html")
	} else if uri == "/analysis/user/" {
		http.ServeFile(w, r, "./pages/analysis/analysis.html")

	} else if uri == "/analysis/user/login/" {
		s := StreamToString(r.Body)
		user_pwd := strings.Split(s, "&")
		if len(user_pwd) != 2 {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		act_user := strings.Split(user_pwd[0], "=")
		act_pwd := strings.Split(user_pwd[1], "=")

		//auth process
		session, err := mgo.DialWithInfo(&c.SearchUser.DbInfo)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		//collection
		collection := session.DB(c.SearchUser.DbInfo.Database).C(c.SearchUser.Collection)
		num, err := collection.Find(bson.M{"username": act_user[1]}).Count()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if num == 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		result := models.CntUser{}
		err = collection.Find(bson.M{"username": act_user[1]}).One(&result)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(act_pwd[1]))
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		//end of auth

		//...get the repo for this user
		session, err = mgo.DialWithInfo(&c.MongoAuth.DbInfo)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		collection = session.DB(c.MongoAuth.DbInfo.Database).C(c.MongoAuth.Collection)
		repo_string := []string{act_user[1]}
		tmp_match := models.ACLEntry{}
		iter := collection.Find(nil).Select(bson.M{"match": 1}).Iter()
		for iter.Next(&tmp_match) {
			if tmp_match.Match.Account == act_user[1] {
				tmp := strings.Split(tmp_match.Match.Name, "/")
				repo_string = append(repo_string, tmp[0])
			}
		}

		session, err = mgo.DialWithInfo(&c.AnalysisConfig.DbInfo)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		//collection
		collection = session.DB(c.AnalysisConfig.DbInfo.Database).C(c.AnalysisConfig.Collection)
		count := 0
		var result_list string
		for _, repo := range repo_string {
			num, err = collection.Find(bson.M{"user": repo}).Count()
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if num == 0 {
				continue
			}

			iter := collection.Find(bson.M{"user": repo}).Iter()
			tmp_repo := models.CntAnalysis{}
			for iter.Next(&tmp_repo) {
				count += 1
				result_list = fmt.Sprintf("%s<p><b>Src:</b>%s <b>Timestamp:</b>%s <b>Action:</b>%s <b>Repository:</b>%s <b>User:</b>%s</p>\r\n", result_list, tmp_repo.Src, tmp_repo.Timestamp, tmp_repo.Action, tmp_repo.Repo, tmp_repo.User)
			}
		}
		if count == 0 {
			http.Error(w, "not record!", http.StatusOK)
			return
		}
		result_list = fmt.Sprintf("<!DOCTYPE html>\r\n<h1>%d item(s) found!</h1>\r\n<h2>Record list below:</h2>\r\n%s</html>\r\n", count, result_list)

		analysis_lock.Lock()
		tmp_file := "./ays_file"
		fout, err := os.Create(tmp_file)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		_, err = fout.WriteString(result_list)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.ServeFile(w, r, tmp_file)
		os.Remove(tmp_file)
		fout.Close()
		analysis_lock.Unlock()
	} else {
		http.NotFound(w, r)
	}
}
