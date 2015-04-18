var app = angular.module('app', []);

app.controller('PersonsController', ['$scope', '$http', function($scope, $http) {
    $scope.persons = [];
    $scope.status = "";

    $http.get('/api/persons')
        .success(function(data) {
            $scope.persons = data;
        })
        .error(function(error) {
            $scope.status = 'Unable to load persons data: ' + error.message;
        });
}]);