




## Close the browser window

1. ReadPings() from Connection get error (err = EOF) when ws.ReadMessage(). That's because browser websocket connection is closed.
2. We terminate ReadPings() and signal to 'unregister' the connection (and close ws on server side).
3. Hub receives the unregister for connection and delete the connection from internal map and close its communication channel 'send'.
4. StartWriting() goroutine detects the send channel is closed and stops itself then signal to 'unregister' the connection (and close ws on server side). 
5. Hub receives again the unregister for connection but find its already unregistered.
 
 
## Integration with Riemann

 In rieman.config.clj
 
    (require '[synygy.riemann.post2dash :refer :all] :reload)
 
    post-dash (post2dash "http://is-memmi.synygy.net:8080/api/event")
 
    dash (where (not (maintenance-mode? event))
                       (where (or (and (any? :org ["Production" "SalesSyn" "PSSyn" "SalesOpty" "SalesSyn"] event)
                                       (not (tagged-any? ["devops" "exception"] event)))
                                  (any? :event ["webmonitor-executionstats"] event))
                              (async-queue! :dash {:queue-size 1000}
                                            (dashboard "/var/log/riemann/dashboard.txt")
                                            (ddashboard (:datomic-uri config))
                                            (changed-state {:init "ok"}
                                                   post-dash)
                                            )))
It will post events to this dashboard by 'change state' which allows event to pass it if it has its current state changed as before. If there is no before state saved in changed-state 
then 'ok' is considered previous state.
                                            
                                            
There must be a file post2dash.clj in /opt/optymyze/riemann/lib/synygy/riemann/post2dash.clj with content:
                                            
    (ns synygy.riemann.post2dash
      (:require [clj-http.client :as client]
            [cheshire.core :as json]
            [clojure.tools.logging :refer [info warn]]))
        
    (defn post2dash [url]
       (info "****  start post2dash for " url)
       (fn [event]
           (let [ data (into {} (remove (comp nil? second) event))
                  data-json (json/generate-string data)]
               ;(info data-json)
               (client/post url {:body data-json :content-type :json}))))
                                                          
## Development

In order to get frequently events to this application from a real Riemann server I use some mock events.
I added some code in riemann.config.clj that transforms ordinary events into alerts by adding status of type 'failure', 'warning', 'critical' or 'ok'.
  
    (require '[synygy.riemann.post2dash :refer :all] :reload)
  
    (where (service #"df-var-log/percent_bytes-free")
          (post2dash "http://is-memmi.synygy.net:8080/api/event"))
          
The script I used to send data is:
       
       (ns synygy.riemann.post2dash
         (:require [clj-http.client :as client]
               [cheshire.core :as json]
               [clojure.tools.logging :refer [info warn]]))
       
       (def states
          ["ok" "critical" "warning" "failure"])
       
       (defn post2dash [url]
          (info "****  start post2dash for " url)
          (fn [event]
              (let [ data (into {} (remove (comp nil? second) event))
                     states (shuffle states)
                     data (assoc data :state (first states))
                     data-json (json/generate-string data)]
                  (client/post url {:body data-json :content-type :json}))))
                  
