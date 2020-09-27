import "./index.css";
import App from "./App";
import React from "react";
import { Provider } from "react-redux";
import { configureStore } from "@reduxjs/toolkit";
import {
  itemUploadFailed,
  itemUploaded,
  itemUploading,
  itemsLoadFailed,
  itemsLoaded,
  itemsLoading,
  loadItems,
  middleware,
  reducer,
  uploadItem,
} from "./store";
import { render } from "react-dom";

const store = configureStore({
  devTools: {
    actionCreators: {
      itemUploadFailed,
      itemUploaded,
      itemUploading,
      itemsLoadFailed,
      itemsLoaded,
      itemsLoading,
      loadItems,
      uploadItem,
    },
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(middleware),
  reducer,
});

store.dispatch(loadItems());

render(
  <React.StrictMode>
    <Provider store={store}>
      <App />
    </Provider>
  </React.StrictMode>,
  document.getElementById("app")
);
