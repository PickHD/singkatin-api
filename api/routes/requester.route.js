const router = require("express").Router();
const { body } = require("express-validator");

//!IMPORT CONTROLLER & MIDDLEWARE
const { createRequestApiKeyHandler } = require("../controllers/requester.controller");
const { createLimiter } = require("../middlewares/rateLimitHandler.middleware");

router.post("/", [createLimiter, body("email").not().isEmpty().isEmail().trim().isLength({ min: 6 })], createRequestApiKeyHandler);

module.exports = router;