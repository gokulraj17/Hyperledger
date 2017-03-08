'use strict';
/*******************************************************************************
 * Copyright (c) 2015 IBM Corp.
 *
 * All rights reserved.
 *
 * Contributors:
 *   David Huffman - Initial implementation
 *   Dale Avery
 *******************************************************************************/
// For logging
var TAG = 'app.js:';

// =====================================================================================================================
// 												Node.js Setup
// =====================================================================================================================
var express = require('express');
var session = require('express-session');
var compression = require('compression');
var serve_static = require('serve-static');
var path = require('path');
var morgan = require('morgan');
var cookieParser = require('cookie-parser');
var bodyParser = require('body-parser');
var http = require('http');
var https = require('https');
var setup = require('./setup');
var cors = require('cors');
var fs = require('fs');
var request = require('request');
var moment = require('moment');
//var restreamer = require('connect-restreamer');

// =====================================================================================================================
// 												Express Setup
// =====================================================================================================================
// Create the Express app that will process incoming requests to our web server.
console.log(TAG, 'Configuring Express app');
var app = express();
//app.set('views', path.join(__dirname, 'views'));
//app.set('view engine', 'pug');
//app.engine('.html', require('pug').__express);
app.set('view engine', 'html');
app.engine('html', require('ejs').renderFile);
app.use(compression());
app.use(morgan('dev'));
app.use(bodyParser.json());
app.use(bodyParser.urlencoded({extended: true}));
app.use(cookieParser());
//app.use(restreamer());

// Create a static folder to serve up the CSS and JS for the demo.  These images shouldn't change very often, so we
// can set longer cache limits for them.
app.use(serve_static(path.join(__dirname, 'public'), {maxAge: '1d', setHeaders: setCustomCC})); // 1 day cache
function setCustomCC(res, path) {
    // 30 days cache
    if (serve_static.mime.lookup(path) === 'image/jpeg') res.setHeader('Cache-Control', 'public, max-age=2592000');
    else if (serve_static.mime.lookup(path) === 'image/png') res.setHeader('Cache-Control', 'public, max-age=2592000');
    else if (serve_static.mime.lookup(path) === 'image/x-icon') res.setHeader('Cache-Control', 'public, max-age=2592000');
}

app.use('/public/css',express.static(__dirname + '/public/css'));
app.use('/public/imgs',express.static(__dirname + '/public/imgs'));
app.use('/public/js', express.static(__dirname + '/public/js'));
app.use('/views', express.static(__dirname + '/views'));

// Use a session to track how many requests we receive from a client (See below)
app.use(session({secret: 'Somethignsomething1234!test', resave: true, saveUninitialized: true}));

// Enable CORS preflight across the board so browser will let the app make REST requests
app.options('*', cors());
app.use(cors());

// Attach useful things to the request
app.use(function (req, res, next) {
    console.log('----------------------------------------- incoming request -----------------------------------------');
    // Create a bag for passing information back to the client
    req.bag = {};
    req.session.count = req.session.count + 1;
    req.bag.session = req.session;
    next();
});

//Cloudant DB connection
var dbCredentials = {
	dbName: 'cm_poc_auth_db'
};
//function initDBConnection() {
	
	var Cloudant = require('cloudant');
	
	// Initialize the library with my account.
	var cloudant = Cloudant({ url: "https://3e2df5bc-9cc9-415e-b0d8-c08f880d830f-bluemix.cloudant.com"}, function(err, data) {
	  console.log(err,data);
	  //console.log("here");
	});
	
	var db = cloudant.db.use("cm_poc_auth_db");

app.post('/authenticate', function (req, res) {
	var username , password, role;
	//console.log("Got request "+req.body);
	try{
		username = req.body.username;
		password = req.body.password;
		role = req.body.role;
		console.log(username+" "+password+" "+role );
		db.list({'include_docs':true}, function(err, doc) {
			if(err){
				console.log("db error "+err.stack);
				 res.json({ success: false, message: err });
				 return;
			}
			console.log("Inside db "+JSON.stringify(doc));
			console.log("Inside db rows"+JSON.stringify(doc.rows));
			//console.log("Inside db rows id "+doc.rows[0].id);
			//console.log("Inside db rows id user_name "+doc.rows[0].user_name);
			//console.log("Inside db rows"+doc);
			var ulist = doc.rows;
			console.log("ulist "+ulist);
			console.log("ulist "+JSON.stringify(ulist));
			var foundCredentials = false;
			for(var i=0; i< ulist.length; i++){
				console.log("Inside for loop "+ulist[i].doc.user_name);
				if(ulist[i].doc.user_name == username){
					console.log(ulist[i].doc.user_role+" "+role);
					if (ulist[i].doc.user_role === role){
						foundCredentials = true;
						if(ulist[i].doc.user_pswd === password){
							console.log("Login is successful for: " + username);
							
							//Add session parameters
					        //request.session.user = ulist[i].gmwID;
					        //request.session.message = '';
					        //request.session.tabName = 'Home';
					        res.json({success: true, message: ulist[i].doc.user_id });
							//response.render("subscriber.html");
							return;
						}
					}
				}
			}
			
			//If the user was not found
			var error_message = "Error: Invalid password!";
			if(foundCredentials == false) {
				error_message = "Error: Could not find the user!";
			}
			console.log("Failed response "+error_message);
			res.json({'success': false, 'message': error_message});
		});
	}catch(er){
		response.statusCode = 400;
		console.log("Authentication failed. Exception occured!!! "+er.message);
		return res.json({ success: false, message: er.message });
	}
});

app.post('/placeFIOrders', function (req, res) {
	console.log("Got request "+JSON.stringify(req.body));
	var orderTime = moment().format("YYYY-MM-DDTHH:mm:ssZ");
	try{
		var order_count = req.body.fiorders.length;
		console.log(order_count);
		for(i = 0; i < order_count; i++){
			req.body.fiorders[i].Quantity = parseInt(req.body.fiorders[i].Quantity);
			req.body.fiorders[i].LimitPrice = parseFloat(req.body.fiorders[i].LimitPrice);
			req.body.fiorders[i]["creationDate"] = orderTime;
			console.log(typeof(req.body.fiorders[i].Quantity)+" "+typeof(req.body.fiorders[i].LimitPrice));
			console.log("creation time is "+req.body.fiorders[i].creationDate);
		}
		console.log(typeof(req.body.fiorders));
		var orders = JSON.stringify(req.body.fiorders);
		console.log(typeof(orders));
		var request_obj = {
				method: 'POST', // do POST
			    url: 'https://505db20ccb684d54a65f14af73f0b7c9-vp0.us.blockchain.ibm.com:5001/chaincode', // here only the domain name
			    json: {
			        "jsonrpc": "2.0",
			        "method": "invoke",
			        "params": {
			            "type": 1,
			            "chaincodeID": {
			                "name": "bf67032e323b0c51984eefa124825fa93d79c5b8d6337308cd58b2d80178d428"
			            },
			            "ctorMsg": {
			              "function": "createOrdersByFI",
			              "args": [
			                   "createOrdersByFI", 
			                   orders
			                
			                                    ]
			            },
			            "secureContext": "user_type1_0"
			        },
			        "id": 0
			    }
			};
		
		request.post(request_obj, function (error, response, body) {
		    if (error) {
		        console.log(error);
		        res.json({success: false, message: error });
		        return;
		    }
		    console.log(response.statusCode);
		    if (response.statusCode == 200) {
		        console.log("Done") // Show the HTML for the Google homepage.
		        console.log(body);
		        res.json({success: true, message: 'Order placed successfully' });
		    } else {
		    	console.log('Some error occurred');
		    	res.json({success: false, message: 'Some error occurred' });
		    }

		});
	}catch(er){
		response.statusCode = 400;
		console.log("Exception occured while placing orders!!! "+er.message);
		return res.json({ success: false, message: er.message });
	}
});

app.get('/getFIOrders', function (req, res) {
	var fiid = req.param('fiid');
	var status = req.param('status');
	console.log(fiid+" "+status);
	try{
		var request_obj = {
			    method: 'POST', // do POST
			    url: 'https://505db20ccb684d54a65f14af73f0b7c9-vp0.us.blockchain.ibm.com:5001/chaincode', // here only the domain name
			    json: {
			        "jsonrpc": "2.0",
			        "method": "query",
			        "params": {
			            "type": 1,
			            "chaincodeID": {
			                "name": "bf67032e323b0c51984eefa124825fa93d79c5b8d6337308cd58b2d80178d428"
			            },
			            "ctorMsg": {
			              "function": "getAllOrdersForFIMap",
			              "args": [
			              ]
			            },
			            "secureContext": "user_type1_0"
			        },
			        "id": 0
			    }
		}
		
		request.post(request_obj, function (error, response, body) {
		    if (error) {
		        console.log(error);
		        res.json({success: false, message: error });
		        return;
		    }
		    console.log(response.statusCode);
		    if (response.statusCode == 200) {
		        console.log("Fetched") // Show the HTML for the Google homepage.
		        console.log(body);
		        //res.json({success: true, message: body });
		        res.send(body);
		    } else {
		    	console.log('Some error occurred');
		    	res.json({success: false, message: 'Some error occurred' });
		    }

		});
	}catch(er){
		response.statusCode = 400;
		console.log("Exception occured while fetching orders!!! "+er.message);
		return res.json({ success: false, message: er.message });
	}
});

app.get('/getBrOrders', function (req, res) {
	var brokerid = req.param('BrokerID');
	var status = req.param('status');
	console.log(brokerid+" "+status);
	try{
		var request_obj = {
			    method: 'POST', // do POST
			    url: 'https://505db20ccb684d54a65f14af73f0b7c9-vp1.us.blockchain.ibm.com:5001/chaincode', // here only the domain name
			    json: {
			        "jsonrpc": "2.0",
			        "method": "query",
			        "params": {
			            "type": 1,
			            "chaincodeID": {
			                "name": "bf67032e323b0c51984eefa124825fa93d79c5b8d6337308cd58b2d80178d428"
			            },
			            "ctorMsg": {
			              "function": "getAllOrdersForBrokerBasedOnStatus",
			              "args": [
			                    
			            	  brokerid, status
			                
			                                    ]
			            },
			            "secureContext": "user_type1_1"
			        },
			        "id": 0
			    }
		}
		
		request.post(request_obj, function (error, response, body) {
		    if (error) {
		        console.log(error);
		        res.json({success: false, message: error });
		        return;
		    }
		    console.log(response.statusCode);
		    if (response.statusCode == 200) {
		        console.log("Broker orders Fetched") // Show the HTML for the Google homepage.
		        console.log(body);
		        //res.json({success: true, message: body });
		        res.send(body);
		    } else {
		    	console.log('Some error occurred');
		    	res.json({success: false, message: 'Some error occurred' });
		    }

		});
	}catch(er){
		response.statusCode = 400;
		console.log("Exception occured while fetching orders!!! "+er.message);
		return res.json({ success: false, message: er.message });
	}
});

app.get('/getSeOrders', function (req, res) {
	var status = req.param('status');
	console.log("--->Status : " +status);
	try{
		var request_obj = {
			    method: 'POST', // do POST
			    url: 'https://505db20ccb684d54a65f14af73f0b7c9-vp3.us.blockchain.ibm.com:5001/chaincode', // here only the domain name
			    json: {
			        "jsonrpc": "2.0",
			        "method": "query",
			        "params": {
			            "type": 1,
			            "chaincodeID": {
			                "name": "bf67032e323b0c51984eefa124825fa93d79c5b8d6337308cd58b2d80178d428"
			            },
			            "ctorMsg": {
			              "function": "getAllOrdersForSEBasedOnStatus",
			              "args": [
			                    
			                    "toBeSettled"
			                    ]
			            },
			            "secureContext": "user_type1_3"
			        },
			        "id": 0
			    }
		}
		
		request.post(request_obj, function (error, response, body) {
		    if (error) {
		        console.log(error);
		        res.json({success: false, message: error });
		        return;
		    }
		    console.log(response.statusCode);
		    if (response.statusCode == 200) {
		        console.log("Broker orders Fetched") // Show the HTML for the Google homepage.
		        console.log(body);
		        //res.json({success: true, message: body });
		        res.send(body);
		    } else {
		    	console.log('Some error occurred');
		    	res.json({success: false, message: 'Some error occurred' });
		    }

		});
	}catch(er){
		response.statusCode = 400;
		console.log("Exception occured while fetching orders!!! "+er.message);
		return res.json({ success: false, message: er.message });
	}
});

app.put('/ConfirmOrders', function (req, res) {
	console.log("Got request for confirmation "+JSON.stringify(req.body));
	var brOrders = req.body.brorders;
	try{
		var request_obj = {
			    method: 'POST', // do POST
			    url: 'https://505db20ccb684d54a65f14af73f0b7c9-vp1.us.blockchain.ibm.com:5001/chaincode', // here only the domain name
			    json: {
			        "jsonrpc": "2.0",
			        "method": "invoke",
			        "params": {
			            "type": 1,
			            "chaincodeID": {
			                "name": "bf67032e323b0c51984eefa124825fa93d79c5b8d6337308cd58b2d80178d428"
			            },
			            "ctorMsg": {
			              "function": "processFIOrdersForConfirmationBySE",
			              "args": [
			            	  "processFIOrdersForConfirmationBySE",
			            	  brOrders
			                
			                                    ]
			            },
			            "secureContext": "user_type1_1"
			        },
			        "id": 0
			    }
		}
		
		request.post(request_obj, function (error, response, body) {
		    if (error) {
		        console.log(error);
		        res.json({success: false, message: error });
		        return;
		    }
		    console.log(response.statusCode);
		    if (response.statusCode == 200) {
		        console.log("Broker orders confirmed") // Show the HTML for the Google homepage.
		        console.log(body);
		        //res.json({success: true, message: body });
		        res.json({success: true, message: 'Orders Confirmed Successfully' });
		    } else {
		    	console.log('Some error occurred');
		    	res.json({success: false, message: 'Some error occurred' });
		    }

		});
	}catch(er){
		response.statusCode = 400;
		console.log("Exception occured while fetching orders!!! "+er.message);
		return res.json({ success: false, message: er.message });
	}
});


app.put('/SettleTrades', function (req, res) {
	console.log("Got request for settlement "+JSON.stringify(req.body));
	var settleTime = moment().format("YYYY-MM-DDTHH:mm:ssZ");
	console.log("Time for Settlement : " + settleTime );
	try{
		var request_obj = {
			    method: 'POST', // do POST
			    url: 'https://505db20ccb684d54a65f14af73f0b7c9-vp3.us.blockchain.ibm.com:5001/chaincode', // here only the domain name
			    json: {
			        "jsonrpc": "2.0",
			        "method": "invoke",
			        "params": {
			            "type": 1,
			            "chaincodeID": {
			                "name": "bf67032e323b0c51984eefa124825fa93d79c5b8d6337308cd58b2d80178d428"
			            },
			            "ctorMsg": {
			              "function": "clearAndSettleTrades",
			              "args": [
			            	  "clearAndSettleTrades",
			            	  settleTime
			                
			                                    ]
			            },
			            "secureContext": "user_type1_3"
			        },
			        "id": 0
			    }
		}
		
		request.post(request_obj, function (error, response, body) {
		    if (error) {
		        console.log(error);
		        res.json({success: false, message: error });
		        return;
		    }
		    console.log(response.statusCode);
		    if (response.statusCode == 200) {
		        console.log("Trade settled successfully") // Show the HTML for the Google homepage.
		        console.log(body);
		        //res.json({success: true, message: body });
		        res.json({success: true, message: 'Trade settled successfully' });
		    } else {
		    	console.log('Some error occurred while settling the trade');
		    	res.json({success: false, message: 'Some error occurred while settling the trade' });
		    }

		});
	}catch(er){
		response.statusCode = 400;
		console.log("Exception occured while settling the trade!!! "+er.message);
		return res.json({ success: false, message: er.message });
	}
});



app.get('/viewCustOrders', function (req, res) {
	var custodianId = req.param('custodian');
	try{
		var request_obj = {
			    method: 'POST', // do POST
			    url: 'https://505db20ccb684d54a65f14af73f0b7c9-vp0.us.blockchain.ibm.com:5001/chaincode', // here only the domain name
			    json: {
			        "jsonrpc": "2.0",
			        "method": "query",
			        "params": {
			            "type": 1,
			            "chaincodeID": {
			                "name": "bf67032e323b0c51984eefa124825fa93d79c5b8d6337308cd58b2d80178d428"
			            },
			            "ctorMsg": {
			              "function": "getAllOrdersForCustodianBasedOnStatus",
			              "args": [
			            	  custodianId
			              ]
			            },
			            "secureContext": "user_type1_0"
			        },
			        "id": 0
			    }
		}
		
		request.post(request_obj, function (error, response, body) {
		    if (error) {
		        console.log(error);
		        res.json({success: false, message: error });
		        return;
		    }
		    console.log(response.statusCode);
		    if (response.statusCode == 200) {
		        console.log("Fetched") // Show the HTML for the Google homepage.
		        console.log(body);
		        //res.json({success: true, message: body });
		        res.send(body);
		    } else {
		    	console.log('Some error occurred');
		    	res.json({success: false, message: 'Some error occurred' });
		    }

		});
	}catch(er){
		response.statusCode = 400;
		console.log("Exception occured while fetching orders!!! "+er.message);
		return res.json({ success: false, message: er.message });
	}
});

app.get('/viewHoldings', function (req, res) {
	var fiid = req.param('fiid');

	try{
		var request_obj = {
			    method: 'POST', // do POST
			    url: 'https://505db20ccb684d54a65f14af73f0b7c9-vp0.us.blockchain.ibm.com:5001/chaincode', // here only the domain name
			    json: {
			        "jsonrpc": "2.0",
			        "method": "query",
			        "params": {
			            "type": 1,
			            "chaincodeID": {
			                "name": "bf67032e323b0c51984eefa124825fa93d79c5b8d6337308cd58b2d80178d428"
			            },
			            "ctorMsg": {
			              "function": "getAllHoldingsForFI",
			              "args": [
			            	  fiid
			              ]
			            },
			            "secureContext": "user_type1_0"
			        },
			        "id": 0
			    }
		}
		
		request.post(request_obj, function (error, response, body) {
		    if (error) {
		        console.log(error);
		        res.json({success: false, message: error });
		        return;
		    }
		    console.log(response.statusCode);
		    if (response.statusCode == 200) {
		        console.log("Fetched") // Show the HTML for the Google homepage.
		        console.log(body);
		        //res.json({success: true, message: body });
		        res.send(body);
		    } else {
		    	console.log('Some error occurred');
		    	res.json({success: false, message: 'Some error occurred' });
		    }

		});
	}catch(er){
		response.statusCode = 400;
		console.log("Exception occured while fetching orders!!! "+er.message);
		return res.json({ success: false, message: er.message });
	}
});

app.get('/viewTransactionBasedOnStock', function (req, res) {
	var fiid = req.param('fiid');
	var stock = req.param('stock');
	try{
		var request_obj = {
			    method: 'POST', // do POST
			    url: 'https://505db20ccb684d54a65f14af73f0b7c9-vp0.us.blockchain.ibm.com:5001/chaincode', // here only the domain name
			    json: {
			        "jsonrpc": "2.0",
			        "method": "query",
			        "params": {
			            "type": 1,
			            "chaincodeID": {
			                "name": "bf67032e323b0c51984eefa124825fa93d79c5b8d6337308cd58b2d80178d428"
			            },
			            "ctorMsg": {
			              "function": "getAllTransactionsForFIBasedOnStockId",
			              "args": [
			            	  fiid, stock
			              ]
			            },
			            "secureContext": "user_type1_0"
			        },
			        "id": 0
			    }
		}
		
		request.post(request_obj, function (error, response, body) {
		    if (error) {
		        console.log(error);
		        res.json({success: false, message: error });
		        return;
		    }
		    console.log(response.statusCode);
		    if (response.statusCode == 200) {
		        console.log("Fetched transactions for stock") // Show the HTML for the Google homepage.
		        console.log(body);
		        //res.json({success: true, message: body });
		        res.send(body);
		    } else {
		    	console.log('Some error occurred');
		    	res.json({success: false, message: 'Some error occurred' });
		    }

		});
	}catch(er){
		response.statusCode = 400;
		console.log("Exception occured while fetching transactions for stock!!! "+er.message);
		return res.json({ success: false, message: er.message });
	}
});

// This router will serve up our pages and API calls.
var router = require('./routes/site_router');
app.use('/', router);
//app.use('/', express.static(__dirname + '/index.html'));

// If the request is not process by this point, their are 2 possibilities:
// 1. We don't have a route for handling the request
app.use(function (req, res, next) {
    var err = new Error('Not Found');
    err.status = 404;
    next(err);
});

// 2. Something else went wrong
app.use(function (err, req, res, next) {		// = development error handler, print stack trace
    console.log(TAG, 'Error Handler -', req.url);
    var errorCode = err.status || 500;
    res.status(errorCode);
    req.bag.error = {msg: err.stack, status: errorCode};
    if (req.bag.error.status == 404) req.bag.error.msg = 'Sorry, I cannot locate that file';
    res.render('template/error', {bag: req.bag});
});

// =====================================================================================================================
// 												Launch Webserver
// =====================================================================================================================
// Start the web server using our express app to handle requests
var host = setup.SERVER.HOST;
var port = setup.SERVER.PORT;
console.log(TAG, 'Staring http server on: ' + host + ':' + port);
var server = http.createServer(app).listen(port, function () {
    console.log(TAG, 'Server Up - ' + host + ':' + port);
});

// Some setting that we've found make our life easier
process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';
server.timeout = 240000;

// Track application bluemix deployments.  All we're tracking is number of deployments.
console.log(TAG, '---- Tracking Deployment');
require('cf-deployment-tracker-client').track();

// =====================================================================================================================
// 												Network credentials
// =====================================================================================================================
// Credentials are first loaded from a file, then overwritten with whatever is found in VCAP

// The network credentials that we will need
var peerURLs = [];
var caURL = null;
var users = null;
var peerHosts = [];

// Load credentials from a file
try {
    console.log(TAG, 'Attempting to read hardcoded network credentials...');
    var manual = JSON.parse(fs.readFileSync('mycreds.json', 'utf8'));

    // Sometimes the credentials from Bluemix are wrapped, sometimes not.
    if (manual.credentials) {
        manual = manual.credentials;
    }

    var peers = manual.peers;
    for (var i in peers) {
        peerURLs.push('grpcs://' + peers[i].discovery_host + ':' + peers[i].discovery_port);
        peerHosts.push('' + peers[i].discovery_host);
    }
    var ca = manual.ca;
    for (var i in ca) {
        caURL = 'grpcs://' + ca[i].url;
    }
    console.log(TAG, 'loading hardcoded peers');
    users = null;																			//users are only found if security is on
    if (manual.users) users = manual.users;
    console.log(TAG, 'loading hardcoded users');
}
catch (e) {
    console.log(TAG, 'Error - could not find hardcoded peers/users, this is okay if running in bluemix');
}

if (process.env.VCAP_SERVICES) {
    //load from vcap, search for service, 1 of the 3 should be found...
    var servicesObject = JSON.parse(process.env.VCAP_SERVICES);
    for (var i in servicesObject) {
        if (i.indexOf('ibm-blockchain') >= 0) {											//looks close enough
            if (servicesObject[i][0].credentials.error) {
                console.log(TAG, '!\n!\n! Error from Bluemix: \n', servicesObject[i][0].credentials.error, '!\n!\n');
                peers = null;
                users = null;
                process.error = {
                    type: 'network',
                    msg: 'Due to overwhelming demand the IBM Blockchain Network service is at maximum capacity.  Please try recreating this service at a later date.'
                };
            }
            if (servicesObject[i][0].credentials && servicesObject[i][0].credentials.peers) {
                console.log('overwritting peers, loading from a vcap service: ', i);
                peers = servicesObject[i][0].credentials.peers;
                peerURLs = [];
                peerHosts = [];
                for (var j in peers) {
                    peerURLs.push('grpcs://' + peers[j].discovery_host + ':' + peers[j].discovery_port);
                    peerHosts.push('' + peers[j].discovery_host);
                }
                if (servicesObject[i][0].credentials.ca) {
                    console.log('overwritting ca, loading from a vcap service: ', i);
                    ca = servicesObject[i][0].credentials.ca;
                    for (var z in ca) {
                        caURL = 'grpcs://' + ca[z].discovery_host + ':' + ca[z].discovery_port;
                    }
                    if (servicesObject[i][0].credentials.users) {
                        console.log('overwritting users, loading from a vcap service: ', i);
                        users = servicesObject[i][0].credentials.users;
                        //TODO extract registrar from users once user list has been updated to new SDK
                    }
                    else users = null;													//no security	
                }
                else ca = null;
                break;
            }
        }
    }
}

// =====================================================================================================================
// 												Blockchain Setup
// =====================================================================================================================
/*console.log(TAG, 'configuring the chain object and its dependencies');

// Things that require the network to be set up
var user_manager = require('./utils/users');
var chaincode_ops = require('./utils/chaincode_ops');
var part2 = require('./utils/ws_part2');

// Keep the keyValStore in the project directory
var keyValStoreDir = __dirname + '/keyValStore';

// Connecting to TLS enabled peers requires a certificate
var certificate = fs.readFileSync('us.blockchain.ibm.com.cert'); // TODO should download using service credentials

// Deploying chaincode requires us to know a path to a certificate on the peers :(
var certificate_path = '/certs/peer/cert.pem'; // TODO this should be available in the service credentials

// Search for chaincode under <project_dir>/src/
process.env.GOPATH = __dirname;

// Create a hfc chain object and deploy our chaincode
var chain_setup = require('./utils/chain_setup');
chain_setup.setupChain(keyValStoreDir, users, peerURLs, caURL, certificate, certificate_path,
    function (error, chain, chaincodeID) {

        if(error) {
            console.log(TAG, 'Chain setup failed:', error);
            throw error;
        }

        // Setup anyone who needs the chain object or the chaincode
        user_manager.setup(chain);

        // Operation involving chaincode in this app should use this object.
        var cpChaincode = new chaincode_ops.CPChaincode(chain, chaincodeID);

        part2.setup(peers, cpChaincode);
        router.setup_helpers(cpChaincode);

        // Now that the chain is ready, start the web socket server so clients can use the demo.
        start_websocket_server();
    });*/

// =====================================================================================================================
// 											WebSocket Communication Madness
// =====================================================================================================================
var ws = require('ws');
var wss = {};

function start_websocket_server(error, d) {
    if (error != null) {
        //look at tutorial_part1.md in the trouble shooting section for help
        console.log('! looks like the final configuration failed, holding off on the starting the socket\n', error);
        if (!process.error) process.error = {type: 'deploy', msg: error.message};
    }
    else {
        console.log('------------------------------------------ Websocket Up ------------------------------------------');
        wss = new ws.Server({server: server});												//start the websocket now
        wss.on('connection', function connection(ws) {
            ws.on('message', function incoming(message) {
                console.log('received ws msg:', message);
                var data = JSON.parse(message);
                part2.process_msg(ws, data);

            });
            ws.on('close', function () {
            });
        });

        // This makes it easier to contact our clients
        wss.broadcast = function broadcast(data) {
            wss.clients.forEach(function each(client) {
                try {
                    data.v = '2';
                    client.send(JSON.stringify(data));
                }
                catch (e) {
                    console.log('error broadcast ws', e);
                }
            });
        };

        // Monitor chain's blockheight and pass it along to clients.
        setInterval(function () {
            var options = {
                host: peers[0].api_host,
                port: peers[0].api_port,
                path: '/chain',
                method: 'GET'
            };

            function success(statusCode, headers, resp) {
                resp = JSON.parse(resp);
                if (resp && resp.height) {
                    wss.broadcast({msg: 'reset'});
                }
            }

            function failure(statusCode, headers, msg) {
                // Don't broadcast failures to clients, just log them
                console.error('chainstats failure: (' +
                    'status code: ' + statusCode +
                    '\n  headers: ' + headers +
                    '\n  message: ' + msg + ')');
            }

            var request = https.request(options, function (resp) {
                var str = '', chunks = 0;
                resp.setEncoding('utf8');
                resp.on('data', function (chunk) {                                                            //merge chunks of request
                    str += chunk;
                    chunks++;
                });
                resp.on('end', function () {
                    if (resp.statusCode == 204 || resp.statusCode >= 200 && resp.statusCode <= 399) {
                        success(resp.statusCode, resp.headers, str);
                    }
                    else {
                        failure(resp.statusCode, resp.headers, str);
                    }
                });
            });

            request.on('error', function (e) {
                failure(500, null, e);
            });

            request.setTimeout(20000);
            request.on('timeout', function () {                                                                //handle time out event
                failure(408, null, 'Request timed out');
            });

            request.end();
        }, 5000);
    }
}
