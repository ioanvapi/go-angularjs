var app = angular.module('app', []);

app.controller('DashboardCtrl', ['$scope', '$http', function ($scope, $http) {
    $scope.events = [];

    $http.get('/api/events')
        .success(function(data) {
            console.log("Data : '" + data + "'")
            if (data) {
                $scope.events = eval(data);
            }
        })
        .error(function(data) {
            console.log(data)
        });

    var conn = new WebSocket("ws://localhost:8080/ws");

    conn.onmessage = function(message) {
        $scope.$apply(function() {
            $scope.events = eval(message.data);
        });
    };

    conn.onclose = function (message) {
       console.log("Websocket connection closed.");
    };

    conn.onopen = function(message) {
        console.log("Websocket connection opened.");
    };

}]);