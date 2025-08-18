const serverPath = import.meta.env.VITE_API_BASE_URL

const paths = {
  apps: `${serverPath}/function_apps`,
  functionCompositions: `${serverPath}/function_compositions`,
  deployments: `${serverPath}/deployments`,
  metrics: `${serverPath}/metrics`,  
}

export default paths;