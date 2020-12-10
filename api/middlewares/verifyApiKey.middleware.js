const { Token } = require("../models");

const verifyApiKey = async (req, res, next) => {
    try {
        const apiKey = req.header("X-API-KEY");
        const getToken = await Token.findOne({
            where: {
                token: apiKey
            }
        });
        if (apiKey === "" || apiKey === undefined) {
            res.statusCode = 400;
            const badRequest = new Error("No Token Provided.");
            return next(badRequest);
        }
        if (getToken !== apiKey || !getToken) {
            res.statusCode = 401;
            const unauthorized = new Error("Key Invalid. Maybe Expired");
            return next(unauthorized);
        }
    } catch (e) {
        res.statusCode = 500;
        return next(e);
    }

};

module.exports = verifyApiKey;