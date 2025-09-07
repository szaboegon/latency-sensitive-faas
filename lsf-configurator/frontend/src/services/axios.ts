import axios from "axios";
import applyCaseMiddleware from "axios-case-converter";

const axiosInstance = applyCaseMiddleware(
  axios.create()
);

export default axiosInstance;
