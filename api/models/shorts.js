"use strict";
const { nanoid } = require("nanoid");
const {
    Model
} = require("sequelize");
module.exports = (sequelize, DataTypes) => {
    class Shorts extends Model {
        /**
         * Helper method for defining associations.
         * This method is not a part of Sequelize lifecycle.
         * The `models/index` file will call this method automatically.
         */
        static associate(models) {
            // define association here
        }
    }
    Shorts.init({
        full_url: {
            type: DataTypes.STRING,
            allowNull: false
        },
        short_url: {
            type: DataTypes.STRING,
            allowNull: false,
            defaultValue:nanoid(7)
        },
        visited: {
            type: DataTypes.INTEGER,
            allowNull: false,
            defaultValue: 0
        },
    }, {
        sequelize,
        modelName: "Shorts",
    });
    return Shorts;
};