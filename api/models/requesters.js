"use strict";
const {
    Model
} = require("sequelize");
module.exports = (sequelize, DataTypes) => {
    class Requesters extends Model {
        /**
         * Helper method for defining associations.
         * This method is not a part of Sequelize lifecycle.
         * The `models/index` file will call this method automatically.
         */
        static associate(models) {
            // define association here
        }
    }
    Requesters.init({
        email: {
            type: DataTypes.STRING,
            isEmail: true,
            allowNull: false,
        },
    }, {
        sequelize,
        modelName: "Requesters",
    });
    return Requesters;
};