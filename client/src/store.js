import { createAction, createReducer } from "@reduxjs/toolkit";

export const itemsLoading = createAction("itemsLoading");
export const itemsLoaded = createAction("itemsLoaded");
export const itemsLoadFailed = createAction("itemsLoadFailed");

export const itemUploading = createAction("itemUploading");
export const itemUploaded = createAction("itemUploaded");
export const itemUploadFailed = createAction("itemUploadFailed");
export const clearItemUploadError = createAction("clearItemUploadError");

export const loadItems = () => async (dispatch) => {
  dispatch(itemsLoading());

  try {
    const response = await fetch("/api");

    if (response.status !== 200) {
      dispatch(itemsLoadFailed());
      return;
    }

    const items = await response.json();

    dispatch(itemsLoaded(items));
  } catch (error) {
    dispatch(itemsLoadFailed());
  }
};

export const uploadItem = (file) => async (dispatch) => {
  dispatch(itemUploading());

  try {
    const body = new FormData();
    body.set("file", file);

    const response = await fetch("/api", {
      method: "POST",
      body,
    });

    if (response.status === 200) {
      dispatch(itemUploaded());
    } else {
      dispatch(itemUploadFailed());
    }

    dispatch(loadItems());
  } catch (error) {
    dispatch(itemUploadFailed());
  }
};

export const middleware = (store) => {
  let timeout;

  return (next) => (action) => {
    const result = next(action);

    if (action.type === itemUploadFailed.type) {
      if (timeout) {
        clearTimeout(timeout);
      }

      timeout = setTimeout(() => {
        store.dispatch(clearItemUploadError());
      }, 1000);
    }

    return result;
  };
};

export const reducer = createReducer(
  {
    loading: 0,
    items: null,
    error: false,
    uploading: 0,
    uploaded: true,
    uploadError: false,
  },
  {
    [itemsLoading]: (state) => ({
      ...state,
      loading: state.loading + 1,
    }),
    [itemsLoaded]: (state, { payload }) => ({
      ...state,
      loading: state.loading - 1,
      items: payload,
      error: false,
    }),
    [itemsLoadFailed]: (state) => ({
      ...state,
      loading: state.loading - 1,
      error: true,
    }),
    [itemUploading]: (state) => ({
      ...state,
      uploading: state.uploading + 1,
    }),
    [itemUploaded]: (state) => ({
      ...state,
      uploading: state.uploading - 1,
      uploadError: false,
    }),
    [itemUploadFailed]: (state) => ({
      ...state,
      uploading: state.uploading - 1,
      uploadError: true,
    }),
    [clearItemUploadError]: (state) => ({
      ...state,
      uploadError: false,
    }),
  }
);

export const isLoading = (state) => state.loading > 0;
export const getItems = (state) => state.items;
export const isErrored = (state) => state.error;

export const isUploading = (state) => state.uploading > 0;
export const isUploadErrored = (state) => state.uploadError;
