const rateLimit = require("express-rate-limit");

const fetchLimiter = rateLimit({
    windowMs: 15 * 60 * 1000, // 15 minutes
    max: 100, // limit each IP to 100 requests per windowMs
    message: { success: false, err_code: 429, err_message: "Too many fetch requested, please try again after an hour", stack: null }
});
const createLimiter = rateLimit({
    windowMs: 15 * 60 * 1000, // 15 minutes
    max: 3, // start blocking after 3 requests
    message: { success: false, err_code: 429, err_message: "TToo many tokens/request created, please try again after an hour", stack: null }
});

module.exports = {
    fetchLimiter,
    createLimiter
};