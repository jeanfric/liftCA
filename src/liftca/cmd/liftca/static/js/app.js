var microcaApp = angular.module('microcaApp', []);

microcaApp.constant("productName", "muCA");
microcaApp.config(
    ['$routeProvider', 
     function($routeProvider) {
         $routeProvider.
             when('/about', {templateUrl: 'partials/about.html'}).
             when('/contact', {templateUrl: 'partials/contact.html'}).
             when('/ca', {templateUrl: 'partials/ca-list.html', controller: 'caListCtrl'}).
             when('/importca', {templateUrl: 'partials/ca-import.html', controller: 'caImportCtrl'}).
             when('/ca/:caId', {templateUrl: 'partials/ca-detail.html', controller: 'caDetailCtrl'}).
             when('/ca/:caId/cert/:certId', {templateUrl: 'partials/cert.html', controller: 'certDetailCtrl'}).
             otherwise({redirectTo: '/ca'});
     }]);

microcaApp.controller(
    'caListCtrl', 
    function caListCtrl($scope, $http, $location) {

        $scope.predicate = "name";
        $scope.reverse = false;
        $scope.ca = {"visible": true};

        var fetch = function() {
            $http.get('ca').success(function(data) {
                $scope.cas = data;
            });
        }
        fetch();
        
        $scope.generateCA = function(ca) {
            $http
                .post('ca', ca)
                .success(function(data) {
                    $location.path('/ca/' + data.serialNumber)
                });
        }
});

microcaApp.controller(
    'caImportCtrl',
    function caImportCtrl($scope, $http, $location) {
        $scope.ca = {"visible": true};
        $scope.importCA = function(ca) {
             $http
                 .post('ca', ca)
                 .success(function(data) {
                     $location.path('/ca/' + data.serialNumber)
                 });
        }
    }
);

microcaApp.controller(
    'caDetailCtrl', 
    function caDetailCtrl($scope, $routeParams, $http, $location) {

        $scope.predicate = "host";
        $scope.reverse = false;

        var fetch = function() {
            var certs;
            $http.get('ca/' + $routeParams.caId).success(function(data) {
                $scope.ca = data;
            });                          
            $http.get('ca/' + $routeParams.caId + '/cert').success(function(data) {
                $scope.certs = data;
                
                $http.get('ca/' + $routeParams.caId + '/crl').success(function(data) {
                    revokedCerts = data.serialNumbers;
                    _($scope.certs).forEach(function(cert) {
                        cert.isRevoked = _(revokedCerts).contains(cert.serialNumber);
                    });
                }); 
            });
        }
        fetch(); 

        $scope.generateCert = function(cert) {
            $http
                .post('ca/' + $routeParams.caId + '/cert', cert)
                .success(function(data) {
                    $location.path('/ca/' + $routeParams.caId + '/cert/' + data.serialNumber)
                });
        };

    });

microcaApp.filter(
    'toArray',
    function () {
        return function (obj) {
            if (!(obj instanceof Object)) {
                return obj;
            }

            return Object.keys(obj).map(function (key) {
                return Object.defineProperty(obj[key], '$key', {__proto__: null, value: key});
            });
        }
    });

microcaApp.controller(
    'certDetailCtrl', 
    function certDetailCtrl($scope, $routeParams, $http) {
        var fetch = function() {
            $http.get('ca/' + $routeParams.caId).success(function(data) {
                $scope.ca = data;
            });
            $http.get('ca/' + $routeParams.caId + '/crl').success(function(data) {
                $scope.certRevoked = _(data.serialNumbers).contains($routeParams.certId);
            });
            $http.get('ca/' + $routeParams.caId + '/cert/' + $routeParams.certId).success(function(data) {
                var hostRegexp = /^(\w+\.)+\w+$/;
                $scope.isValidLink = false;
                if (hostRegexp.test(data.host)) {
                    $scope.isValidLink = true;
                }

                $scope.cert = data;
            });         
        }
        fetch();
        
        $scope.revokeCert  = function() {
            certToRevoke = { serialNumber: $scope.cert.serialNumber };
            $http
                .post('ca/' + $routeParams.caId + '/crl', certToRevoke)
                .success(function() {
                    fetch();
                });
        };

        $scope.unrevokeCert  = function() {
            $http
                .delete('ca/' + $routeParams.caId + '/crl/' + $scope.cert.serialNumber)
                .success(function() {
                    fetch();
                });
        };
    });
