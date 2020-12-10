const router = require("express").Router();

//!IMPORT CONTROLLER & MIDDLEWARE
const { createRequestApiKeyHandler } = require("../controllers/short.controller");
const { createLimiter } = require("../middlewares/rateLimitHandler.middleware");

// router.post("/", [createLimiter], createRequestApiKeyHandler);

module.exports = router;