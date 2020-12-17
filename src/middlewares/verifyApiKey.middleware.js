const { Token } = require("../models");

const verifyApiKey = async (req, res, next) => {
    try {
        const apiKey = req.header("X-API-KEY");

        if (apiKey === undefined) {
            res.statusCode = 400;
            return next(new Error("No Token Provided."));
        }
        
        const getToken = await Token.findOne({
            where: {
                token: apiKey
            }
        });
        if (!getToken) {
            res.statusCode = 401;
            return next(new Error("Key Invalid. Maybe Expired"));
        }
        next();
    } catch (e) {
        res.statusCode = 500;
        return next(e);
    }

};

module.exports = verifyApiKey;