const serverPath = import.meta.env.VITE_API_BASE_URL;

const functionApps = `${serverPath}/function_apps`;
const functionCompositions = `${serverPath}/function_compositions`;
const deployments = `${serverPath}/deployments`;
const metrics = `${serverPath}/metrics`;

const paths = {
  functionApps: functionApps,
  functionCompositions: functionCompositions,
  deployments: deployments,
  metrics: metrics,

  functionAppsBulkCreate: `${functionApps}/bulk`,
};

export default paths;
