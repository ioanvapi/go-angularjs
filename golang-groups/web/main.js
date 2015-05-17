var app = angular.module('app', []);

app.controller('GroupsCtrl', function($scope, $http) {
    $scope.groups = [];
    $scope.errors = [];

    $http.get('/api/groups')
        .success(function(data){
            console.log(data);
            $scope.groups = data.groups;
            $scope.errors = data.errors;
        })
        .error(function(error) {
            $scope.log(error)
        });

    $scope.log = function(msg) {
        $scope.errors.push(msg);
    }
});