const express = require('express');
const crypto = require('crypto');

const PORT = 8080;

const app = express();

var users = {};

app.get('/newUser', function (req, res) {
	// body...
	var username = req.query.username || '';
	var password = req.query.password || '';

	username = username.replace(/[!@#$%^&*]/g, '');

	if (!username || !password || users.username) {
		return res.sendStatus(400);
	}

	var salt = crypto.randomBytes(128).toString('base64');
	var hash = crypto.pbkdf2Sync(password, salt, 10000, 512);

	users[username] = {
		salt: salt,
		hash: hash
	};
	res.sendStatus(200);
});

app.get('/auth', function (req, res) {
	// body...
	var username = req.query.username || '';
	var password = req.query.password || '';

	username = username.replace(/[!@#$%^&*]/g, '');

	if (!username || !password || !users[username]) {
		return res.sendStatus(400);
	}

	var hash = crypto.pbkdf2Sync(password, users[username].salt, 10000, 512);

	if (users[username].hash.toString() === hash.toString()) {
		res.sendStatus(200);
	} else {
		res.sendStatus(401);
	}
});

app.listen(PORT);