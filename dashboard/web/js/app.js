var app = angular.module('app', ['ngRoute']);


app.config(function ($routeProvider) {
    $routeProvider
        .when('/',
        {
            controller: 'EventsController',
            templateUrl: 'views/events.html'
        })
        .when('/history',
        {
            controller: 'HistoryController',
            templateUrl: 'views/history.html'
        })
        .when('/maintenance',
        {
            controller: 'MaintenanceController',
            templateUrl: 'views/maintenance.html'
        })
        .otherwise({redirectTo: '/'})
});
