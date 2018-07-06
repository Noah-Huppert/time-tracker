"use strict";

const path = require("path");
const webpack = require("webpack");

module.exports = {
	entry: "./src/js/main.js",
	output: {
		path: path.resolve(__dirname, "dist"),
		filename: "main.bundle.js"
	},
	module: {
		rules: [
			{
				test: /\.js$/,
				loader: "babel-loader",
				query: {
					presets: ["es2015"]
				}
			}
		]
	},
	stats: {
		colors: true
	},
	devtool: "source-map"
};
