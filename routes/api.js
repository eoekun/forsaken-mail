/**
 * Created by Hongcai Deng on 2015/12/28.
 */

'use strict';

let express = require('express');
let router = express.Router();
let config = require('../modules/config');

router.get('/', function(req, res) {
  res.end();
});

router.get('/config', function(req, res) {
  res.json({
    host: config.host || req.hostname
  });
});

module.exports = router;
