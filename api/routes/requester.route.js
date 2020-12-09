const router = require("express").Router();

//!IMPORT CONTROLLER & MIDDLEWARE
const { createRequestApiKeyHandler } = require("../controllers/short.controller");

router.post("/", createRequestApiKeyHandler);

module.exports = router;