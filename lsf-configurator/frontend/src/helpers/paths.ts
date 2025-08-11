const serverPath = import.meta.env.VITE_API_BASE_URL

const paths = {
  apps: `${serverPath}/apps`,
  functionCompositions: `${serverPath}/function_compositions`,
  metrics: `${serverPath}/metrics`,

  
}

export default paths;