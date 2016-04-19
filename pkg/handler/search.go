package handler

import (
	"bytes"
	"fmt"
	"github.com/duyanghao/registry-notification-server/config"
	"github.com/duyanghao/registry-notification-server/models"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

var search_lock *sync.Mutex = new(sync.Mutex)

func StreamToString(stream io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.String()
}

// Handle for search request
func ProcessSearch(w http.ResponseWriter, r *http.Request, c *config.Config) {
	uri := r.RequestURI
	if uri == "/search/" {
		http.ServeFile(w, r, "./pages/search/home.html")
	} else if uri == "/search/user/" {
		http.ServeFile(w, r, "./pages/search/repo.html")

	} else if uri == "/search/user/login/" {
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

		//reponse for image repo search request
		//get the repo for this user
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

		session, err = mgo.DialWithInfo(&c.SearchRepo.DbInfo)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		//collection
		collection = session.DB(c.SearchRepo.DbInfo.Database).C(c.SearchRepo.Collection)
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

			var tmp_repo []string
			err := collection.Find(bson.M{"user": repo}).Distinct("repo", &tmp_repo)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			for _, i := range tmp_repo {
				count += 1
				result_list = fmt.Sprintf("%s<p><b>Namespace:</b>%s <b>Repository:</b>%s</p>\r\n", result_list, repo, i)
			}

		}
		if count == 0 {
			http.Error(w, "not record!", http.StatusOK)
			return
		}
		result_list = fmt.Sprintf("<!DOCTYPE html>\r\n<h1>%d item(s) found!</h1>\r\n<h2>Search list below:</h2>\r\n%s</html>\r\n", count, result_list)

		search_lock.Lock()
		tmp_file := "./tmp_file"
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
		search_lock.Unlock()

	} else if uri == "/search/user/repo/" {
		http.ServeFile(w, r, "./pages/search/tag.html")
	} else if uri == "/search/user/repo/login/" {
		s := StreamToString(r.Body)
		user_pwd := strings.Split(s, "&")
		if len(user_pwd) != 3 {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		act_user := strings.Split(user_pwd[0], "=")
		act_pwd := strings.Split(user_pwd[1], "=")
		act_repo := strings.Split(user_pwd[2], "=")

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

		//reponse for image repo search request
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
			//fmt.Printf("%s\n%s\n",tmp_match.Match.Account,tmp_match.Match.Name)
			if tmp_match.Match.Account == act_user[1] {
				tmp := strings.Split(tmp_match.Match.Name, "/")
				repo_string = append(repo_string, tmp[0])
			}
		}

		session, err = mgo.DialWithInfo(&c.SearchRepo.DbInfo)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		//collection
		collection = session.DB(c.SearchRepo.DbInfo.Database).C(c.SearchRepo.Collection)
		count := 0
		var result_list string
		for _, repo := range repo_string {
			num, err = collection.Find(bson.M{"user": repo, "repo": act_repo[1]}).Count()
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if num == 0 {
				continue
			}

			iter := collection.Find(bson.M{"user": repo, "repo": act_repo[1]}).Iter()
			tmp_repo := models.CntRepo{}
			for iter.Next(&tmp_repo) {
				count += 1
				result_list = fmt.Sprintf("%s<p><b>Namespace:</b>%s <b>Repository:</b>%s <b>Tag:</b>%s</p>\r\n", result_list, tmp_repo.User, tmp_repo.Repo, tmp_repo.Tag)
			}
		}
		if count == 0 {
			http.Error(w, "not record!", http.StatusOK)
			return
		}
		result_list = fmt.Sprintf("<!DOCTYPE html>\r\n<h1>%d item(s) found!</h1>\r\n<h2>Search list below:</h2>\r\n%s</html>\r\n", count, result_list)

		search_lock.Lock()
		tmp_file := "./tmp_file"
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
		search_lock.Unlock()

	} else {
		http.NotFound(w, r)
	}

}
