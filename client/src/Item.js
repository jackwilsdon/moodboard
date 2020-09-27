import "./Item.css";
import React from "react";

const Item = ({ id }) => <img className="Item" src={`/api/image/${id}`} />;

export default Item;
