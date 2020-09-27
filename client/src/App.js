import Item from "./Item";
import React, { useEffect } from "react";
import { connect } from "react-redux";
import {
  getItems,
  isErrored,
  isLoading,
  isUploadErrored,
  isUploading,
  uploadItem,
} from "./store";

const Form = ({ uploading, uploadErrored, uploadItem }) => {
  useEffect(() => {
    if (uploading || uploadErrored) {
      return;
    }

    const paste = (event) => {
      if (event.clipboardData.files.length === 0) {
        return;
      }

      uploadItem(event.clipboardData.files[0]);
    };

    document.addEventListener("paste", paste);

    return () => document.removeEventListener("paste", paste);
  }, [uploading, uploadErrored, uploadItem]);

  if (uploading) {
    return "Uploading...";
  }

  if (uploadErrored) {
    return "Upload failed!";
  }

  return (
    <input
      type="file"
      onChange={(event) => {
        uploadItem(event.target.files[0]);
        event.target.value = null;
      }}
    />
  );
};

const App = ({
  loading,
  items,
  errored,
  uploading,
  uploadErrored,
  uploadItem,
}) => {
  if (loading) {
    return "Loading...";
  }

  if (errored) {
    return "Failed to load items!";
  }

  return (
    <>
      <Form
        uploading={uploading}
        uploadErrored={uploadErrored}
        uploadItem={uploadItem}
      />
      {items.map((id) => (
        <Item id={id} key={id} />
      ))}
    </>
  );
};

const mapStateToProps = (state) => ({
  loading: isLoading(state),
  items: getItems(state),
  errored: isErrored(state),
  uploading: isUploading(state),
  uploadErrored: isUploadErrored(state),
});

const mapDispatchToProps = { uploadItem };

export default connect(mapStateToProps, mapDispatchToProps)(App);
