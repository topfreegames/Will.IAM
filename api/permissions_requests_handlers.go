package api

import (
	"fmt"
	"net/http"

	"github.com/topfreegames/Will.IAM/usecases"
	"github.com/topfreegames/extensions/middleware"
)

// func permissionsRequestsCreateHandler(
// 	prsUC usecases.PermissionsRequests,
// ) func(http.ResponseWriter, *http.Request) {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		l := middleware.GetLogger(r.Context())
// 		body, err := ioutil.ReadAll(r.Body)
// 		defer r.Body.Close()
// 		if err != nil {
// 			l.Error(err)
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}
// 		pr := &models.PermissionRequest{}
// 		err = json.Unmarshal(body, pr)
// 		if err != nil {
// 			l.Error(err)
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}
//
// 		saID, _ := getServiceAccountID(r.Context())
// 		has, err := sasUC.WithContext(r.Context()).
// 			HasPermissionString(saID, pr.ToLenderString())
// 		if err != nil {
// 			l.Error(err)
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}
// 		if has {
// 			w.WriteHeader(http.StatusNoContent)
// 			return
// 		}
//
// 		// TODO: check if there's a request with state = Created already NoContent
//
// 		err = psUC.WithContext(r.Context()).CreateRequest(saID, pr)
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}
// 		w.WriteHeader(http.StatusAccepted)
// 	}
// }

func permissionsRequestsListOpenHandler(
	prsUC usecases.PermissionsRequests,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := middleware.GetLogger(r.Context())
		saID, _ := getServiceAccountID(r.Context())
		listOptions, err := buildListOptions(r)
		if err != nil {
			Write(
				w, http.StatusUnprocessableEntity,
				fmt.Sprintf(`{ "error": "%s"  }`, err.Error()),
			)
			return
		}
		prs, count, err := prsUC.WithContext(r.Context()).ListOpenRequestsVisibleTo(listOptions, saID)
		if err != nil {
			l.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		WriteJSON(w, 200, map[string]interface{}{
			"count":   count,
			"results": prs,
		})
	}
}
