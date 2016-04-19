package handler

import (
	"fmt"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/notifications"
	"github.com/duyanghao/registry-notification-server/config"
	"github.com/duyanghao/registry-notification-server/models"
	"gopkg.in/mgo.v2"
	"net/http"
)

//Insert the Manifest pull record into MongoDB analysis_notify
func ProcessPullEvent(w http.ResponseWriter, r *http.Request, e notifications.Event, c *config.Config) error {
	if e.Target.MediaType != schema2.MediaTypeManifest {
		return fmt.Errorf("Wrong event.Target.MediaType: \"%s\". Expected: \"%s\"", e.Target.MediaType, schema2.MediaTypeManifest)
	}
	//create MongoDB Session
	session, err := mgo.DialWithInfo(&c.AnalysisConfig.DbInfo)
	if err != nil {
		return fmt.Errorf("Failed to create Analysis_config MongoDB session: %s", err)
	}
	//collection
	collection := session.DB(c.AnalysisConfig.DbInfo.Database).C(c.AnalysisConfig.Collection)

	repo_tmp := fmt.Sprintf("%s:%s", e.Target.Repository, e.Target.Tag)
	tmp := &models.CntAnalysis{
		Src:       e.Request.Addr,
		Timestamp: e.Timestamp,
		Action:    e.Action,
		Repo:      repo_tmp,
		User:      e.Actor.Name,
	}
	err = collection.Insert(tmp)
	if err != nil {
		return fmt.Errorf("Failed to insert pull record: %s", err)
	}
	fmt.Printf("INFO: pull\n")
	fmt.Printf("INFO: %s \n", e)
	return nil
}
