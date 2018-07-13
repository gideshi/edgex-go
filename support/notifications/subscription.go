/*******************************************************************************
 * Copyright 2018 Dell Technologies Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 *******************************************************************************/

package notifications

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/edgexfoundry/edgex-go/support/notifications/clients"
	"github.com/edgexfoundry/edgex-go/support/notifications/models"
	"github.com/gorilla/mux"
)

const (
	maxExceededString string = "Error, exceeded the max limit as defined in config"
	applicationJson          = "application/json; charset=utf-8"
)

func subscriptionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	slug := vars["slug"]

	switch r.Method {

	// Get all subscriptions
	case http.MethodGet:
		events, err := dbc.Subscriptions()
		if err != nil {
			loggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		// Check max limit
		if len(events) > configuration.ReadMaxLimit {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			loggingClient.Error(maxExceededString)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encode(events, w)
		break

		// Modify (an existing) subscription
	case http.MethodPut:
		var s models.Subscription
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&s)

		// Check if the subscription exists
		s2, err := dbc.SubscriptionBySlug(s.Slug)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			loggingClient.Error(err.Error())
			return
		} else {
			s2 = s
		}

		loggingClient.Info("Updating subscription by slug: " + slug)

		if err = dbc.UpdateSubscription(s2);
			err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			loggingClient.Error(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
		break

	case http.MethodPost:
		var s models.Subscription
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&s)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			loggingClient.Error("Error decoding subscription: " + err.Error())
			return
		}

		loggingClient.Info("Posting Subscription: " + s.String())
		_, err = dbc.AddSubscription(&s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			loggingClient.Error(err.Error())
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(s.Slug))

		break
	}
}

func subscriptionByIDHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	switch r.Method {
	case http.MethodGet:

		s, err := dbc.SubscriptionById(vars["id"])
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(s, w)
	}
}

func subscriptionsBySlugHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	slug := vars["slug"]
	switch r.Method {
	case http.MethodGet:

		s, err := dbc.SubscriptionBySlug(slug)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			loggingClient.Error(err.Error())
			w.Header().Set("Content-Type", applicationJson)
			encode(s, w)
			return
		}

		encode(s, w)
	case http.MethodDelete:
		_, err := dbc.SubscriptionBySlug(slug)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			loggingClient.Error(err.Error())
			return
		}

		loggingClient.Info("Deleting subscription by slug: " + slug)

		if err = dbc.DeleteSubscriptionBySlug(slug); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			loggingClient.Error(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
	}
}

func subscriptionsByCategoriesHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	switch r.Method {
	case http.MethodGet:

		categories := splitVars(vars["categories"])

		s, err := dbc.SubscriptionByCategories(categories)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(s, w)
	}
}

func subscriptionsByLabelsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	switch r.Method {
	case http.MethodGet:

		labels := splitVars(vars["labels"])

		s, err := dbc.SubscriptionByLabels(labels)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(s, w)
	}
}

func subscriptionsByCategoriesLabelsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	switch r.Method {
	case http.MethodGet:

		labels := splitVars(vars["labels"])
		categories := splitVars(vars["categories"])

		s, err := dbc.SubscriptionByCategoriesLabels(categories, labels)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(s, w)
	}
}

func splitVars(vars string) []string {
	return strings.Split(vars, ",")
}

func subscriptionsByReceiverHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	switch r.Method {
	case http.MethodGet:

		s, err := dbc.SubscriptionByReceiver(vars["receiver"])
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(s, w)
	}
}
