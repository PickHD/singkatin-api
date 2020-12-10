const routeNotFound = (req, res, next) => {
    if (!req.route) {
        res.statusCode = 404;
        return next(new Error("Route Not Found."));
    }
    next();
};

module.exports = routeNotFound;