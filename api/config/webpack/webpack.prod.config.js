const { VueLoaderPlugin } = require("vue-loader");
const path = require("path");
const webpack = require("webpack");

module.exports = {
  mode: "production",
  devtool: "source-map",
  entry: {
    application: [
      "./app/javascripts/application.js",
      "./app/assets/stylesheets/application.scss",
    ],
  },
  output: {
    filename: "[name].js",
    sourceMapFilename: "[name].js.map",
    path: path.resolve(__dirname, "../../", "./app/assets/builds"),
  },
  module: {
    rules: [
      {
        test: /\.s[ac]ss$/i,
        use: [
          // Creates `style` nodes from JS strings
          "style-loader",
          // Translates CSS into CommonJS
          "css-loader",
          // Compiles Sass to CSS
          "sass-loader",
        ],
      },
      {
        test: /\.css$/,
        use: [
          "vue-style-loader",
          {
            loader: "css-loader",
            options: { importLoaders: 1 },
          },
          "postcss-loader",
        ],
      },
      {
        test: /\.js$/,
        exclude: /node_modules/,
        use: {
          loader: "babel-loader",
        },
      },
      {
        test: /\.vue$/,
        loader: "vue-loader",
      },
    ],
  },
  resolve: {
    alias: {
      vue$: "vue/dist/vue.runtime.esm-bundler.js",
    },
    extensions: ["*", ".js", ".vue", ".json", "*.map"],
  },
  optimization: {
    chunkIds: "named",

    minimize: true,

    concatenateModules: true,
  },

  plugins: [
    new VueLoaderPlugin(),
    new webpack.optimize.LimitChunkCountPlugin({
      maxChunks: 5,
    }),
    new webpack.DefinePlugin({
      // Drop Options API from bundle
      __VUE_OPTIONS_API__: true,
      __VUE_PROD_DEVTOOLS__: false,
    }),
  ],
};
