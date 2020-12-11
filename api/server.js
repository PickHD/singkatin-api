const express = require("express");
const cors = require("cors");
const helmet = require("helmet");
const morgan = require("morgan");
const apiCache = require("apicache");
require("dotenv").config();

const app = express();
const db = require("./models/index");
const job = require("./helpers/scheduledDelToken.helper");

const apiGateway = require("./routes/index.route");
const errHandler = require("./middlewares/errHandler.middleware");
const routeNotFound = require("./middlewares/routeNotFound.middleware");

(async function setupAPI() {
    try {
        const cache = apiCache.options({
            statusCodes: {
                exclude: [203, 400, 401, 404, 429, 500, 501, 503],
                include: [200],
            }
        }).middleware;

        app.use(express.json());
        app.use(express.urlencoded({ extended: true }));

        app.use(cors());
        app.use(helmet());

        if (process.env.NODE_ENV === "production") {
            app.use(morgan("tiny"));
            app.use(cache("15 min"));
            await db.sequelize.authenticate();
        } else {
            app.use(morgan("dev"));
            db.sequelize.sync({ force: true });
        }

        app.use("/", apiGateway);
        app.use(routeNotFound);
        app.use(errHandler);

        job.start();
    } catch (e) {
        console.error(e);
    }

}());

const server = app.listen(process.env.PORT || 8080, () => {
    //!GET DYNAMIC PORT 
    let port = server.address().port;
    console.log(`Server is running on port :${port}`);
});
