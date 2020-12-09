const routeNotFound = (req, res, next) => {
    if (!req.route) {
        res.statusCode = 404;
        let error = new Error("Route Not Found.");
        return next(error);
    }
};

module.exports = routeNotFound;