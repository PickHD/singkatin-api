const errHandler = (err, req, res, next) => {
    const statusCode = res.statusCode !== 200 ? res.statusCode : 500;
    return res.status(statusCode).json({
        success: false,
        err_code:statusCode,
        err_message: err.message,
        stack: process.env.NODE_ENV === "production" ? "🥞" : err.stack
    });
};

module.exports = errHandler;