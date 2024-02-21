package azuremonitor

//Implemented functional tests using Mockoon tool
//In file azure-http-data-collector-api.json described rules for mock server which will emulate Azure HTTP Data Collector API.
// https://mockoon.com/docs/latest/about/
//Mockoon has capabilities to define rules for checking HTTP requests and sending responses accordingly.
//Proposed rules to validate the `log-type` header and body fields in each record (`message` and `log_type`)
//to ensure that log records are properly formatted, e.g.:
//```
//"rules": [
//            {
// here says that: HTTP request MUST have Header named 'log-type' and with value 'myLogType'
//              "target": "header",
//              "modifier": "log-type",
//              "value": "myLogType",
//              "invert": false,
//              "operator": "equals"
//            },
//            {
// here says that: HTTP request MUST have body in JSON format, with at least one element, that have 'message' key and value "This is my test message"
//              "target": "body",
//              "modifier": "$.[0].message",
//              "value": "This is my test message",
//              "invert": false,
//              "operator": "equals"
//            },
//            {
//              "target": "body",
//              "modifier": "$.[0].log_type",
//              "value": "application",
//              "invert": false,
//              "operator": "equals"
//            }
//          ],
//```
//
//- Set up a mocking environment within a Pod and made it available via a Route to emulate the original API URI format
//- Emulate the original API URI format (`https://<CustomerId>.<Host>/api/logs?api-version=2016-04-01`),
//  including the `<CustomerID>` and `<Host>` components, to accurately replicate the production environment
