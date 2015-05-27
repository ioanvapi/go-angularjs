var app = angular.module('app', []);

app.controller('DashboardCtrl', ['$scope', '$http', function ($scope, $http) {
    $scope.loggedUser = "anonymous";
    $scope.activeEvents = [];
    $scope.ackEvents = [];
    $scope.lastRefresh = new Date().getTime();


    var port = location.port;
    if (port != '') {
        port = ':' + port
    }
    var conn = new WebSocket("ws://" + location.hostname + port + "/ws");

    conn.onmessage = function (message) {
        $scope.$apply(function () {
            var d = JSON.parse(message.data);
            $scope.activeEvents = d.activeEvents;
            $scope.ackEvents = d.ackEvents;
        });
    };

    conn.onclose = function (message) {
        var t = new Date().toLocaleTimeString();
        $scope.disconnectedTime = t;
        $scope.$apply();
        console.log("Websocket connection closed at: ", t);
    };

    conn.onopen = function (message) {
        var t = new Date().toLocaleTimeString();
        $scope.connectedTime = t;
        conn.send('ping');

        console.log("Websocket connection opened at: ", t);

        // send a ping every 30 sec in order to keep the connection alive
        // we must set a bigger value on server (ex: 40)
        setInterval(function () {
            if (conn.readyState == WebSocket.OPEN) {
                conn.send('ping');
            }
        }, 30 * 1000); // when I increased time I got error 'WSARecv tcp [::1]:8080: i/o timeout'
    };


    $scope.ack = function (host, service) {
        var body = JSON.stringify({
            "host": host,
            "service": service,
            "ackUser": "anonymous",
            "ackMessage": "no message"
        });

        $http.post('/api/ackevent', body)
            .success(function (data) {
                console.log('Successful ACK:', data);
            })
            .error(function (data) {
                console.log('Error when ACK:', data);
            });
    };

    // get events to populate Active Events table
    var getActiveEvents = function () {
        $http.get('/api/events')
            .success(function (data) {
                if (data) {
                    $scope.activeEvents = data.activeEvents;
                }
            })
            .error(function (data) {
                console.log(data)
            });
    };

    // get events to populate Acknowledged Events table
    var getAckEvents = function () {
        $http.get('/api/ackevents')
            .success(function (data) {
                //console.log("Ack: " + data);
                if (data) {
                    $scope.ackEvents = data.ackEvents;
                }
            })
            .error(function (data) {
                console.log(data)
            });
    };

    $scope.refresh = function () {
        $scope.lastRefresh = new Date().getTime();
        getActiveEvents();
        getAckEvents();
    };

    getActiveEvents();
    getAckEvents();
}]);