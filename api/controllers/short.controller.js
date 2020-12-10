const { Short } = require("../models");


exports.getAllUrl = async (req, res, next) => {
    try {
    
        //TODO : Create Dynamic Query String For Filter & Searching Rows
        //! const {} = req.query 

        const { count, rows } = await Short.findAndCountAll({ raw: true });

        return res.status(200).json({
            success: true,
            data: {
                count: count,
                rows: rows
            }
        });

    } catch (e) {
        res.statusCode = 500;
        return next(e);
    }

};