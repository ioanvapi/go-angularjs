var app = angular.module('app', []);

app.controller('DashboardCtrl', ['$scope', '$http', function ($scope, $http) {
    $scope.activeEvents = [];
    $scope.ackEvents = [];
    $scope.dash = {};

    // get events to populate Active Events table
    $http.get('/api/events')
        .success(function (data) {
            if (data) {
                $scope.activeEvents = formatTime(data.activeEvents);
            }
        })
        .error(function (data) {
            console.log(data)
        });

    // get events to populate Acknowledged Events table
    $http.get('/api/ackevents')
        .success(function (data) {
            //console.log("Ack: " + data);
            if (data) {
                $scope.ackEvents = formatTime(data.ackEvents);
            }
        })
        .error(function (data) {
            console.log(data)
        });

    var port = location.port;
    if (port != '') {
        port = ':' + port
    }
    var conn = new WebSocket("ws://" + location.hostname + port + "/ws");

    conn.onmessage = function (message) {
        $scope.$apply(function () {
            var d = JSON.parse(message.data);
            $scope.activeEvents = formatTime(d.activeEvents);
            $scope.ackEvents = formatTime(d.ackEvents);
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
        }, 50 * 1000);
    };

    var formatTime = function (events) {
        var data = [];
        for (var i = 0; i < events.length; i++) {
            var event = events[i];
            var t = new Date(event.time * 1000);
            event.timeDisplay = t.toDateString() + " " + t.toTimeString();
            data.push(event);
        }
        return data;
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

}]);