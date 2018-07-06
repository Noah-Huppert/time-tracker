"use strict";

// Require
const path = require("path");
const webpack = require("webpack");

const CleanWebpackPlugin = require("clean-webpack-plugin");
const HtmlWebpackPlugin = require("html-webpack-plugin");

// Constants
const buildDir = path.resolve(__dirname, "dist")

module.exports = {
	// Input
	entry: {
		"js/main.js": "./src/js/main.js"
	},

	// Output
	output: {
		path: buildDir,
		filename: '[name].[chunkhash].js',
	},	

	// Pre-Processors
	module: {
		rules: [
			// Javascript ES6
			{
				test: /\.js$/,
				use: [
					{
						loader: "babel-loader",
						options: {
							presets: ["babel-preset-env"]
						}
					}
				]
			}
		]
	},

	// Plugins
	plugins: [
		// Clean build directory
		new CleanWebpackPlugin([buildDir]),
		
		// HTML generation
		new HtmlWebpackPlugin({
			title: "Time Tracker",
			meta: {
				viewport: "width=device-with, initial-scale=1"
			}
		})
	],

	// Run configuration
	stats: {
		colors: true
	},
	devtool: "source-map"
};
