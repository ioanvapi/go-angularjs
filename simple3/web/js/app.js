var app = angular.module('app', []);

app.controller('PersonsController', ['$scope', '$http', function ($scope, $http) {
    $scope.persons = [];
    $scope.status = "";

    var getPersons = function () {
        $http.get('/api/persons')
            .success(function (data) {
                $scope.debug1 = data;
                $scope.persons = data;
            })
            .error(function (error) {
                $scope.status = 'Unable to load persons data: ' + error.message;
            });
    };

    getPersons();

    $scope.addPerson = function (name) {
        $http.post('/api/person', {Name: name})
            .success(function () {
                $scope.persons.push(name);
                $scope.newName = "";
            })
            .error(function (error) {
                $scope.status = 'Unable to add new person: ' + error.message;
            });
    };

    $scope.deletePerson = function (index) {
        $http.delete('/api/person/' + index)
            .success(function () {
                $scope.persons.splice(index, 1);
            })
            .error(function (error) {
                $scope.status = 'Unable to delete person: ' + error.message;
            });
    };

}]);