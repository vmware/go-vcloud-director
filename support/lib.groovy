// This file contains methods used by Jenkinsfile to support the
// Jenkins pipeline. The contents are loaded after the target
// git repository is checked out.

// Declare global variables so they can be used in all pipeline methods. 
credentialsArray = []
environmentArray = []
temporaryFiles = []

// Process job parameters and determine which credentials or environment
// variables are required for proper processing.
def init() {
    // Determine if a credentials ID has been provided, or if the default should be used.
    def govcdConfigCredentialsID = params.GOVCD_CONFIG_CREDENTIALS_ID
    if (govcdConfigCredentialsID.toLowerCase() == 'default') {
        govcdConfigCredentialsID = env.DEFAULT_GOVCD_CONFIG_CREDENTIALS_ID
    }

    // Check for a parameter containing the GOVCD_CONFIG content
    // Write that to a file if available, or use a Jenkins credential file
    if (env.GOVCD_CONFIG_CONTENTS != null && env.GOVCD_CONFIG_CONTENTS != "") {
        def tmpdir = pwd(tmp:true)
        def vcdPath = "${tmpdir}/jenkins_govcd_config"

        println "Write GOVCD_CONFIG_CONTENTS to ${vcdPath}"
        writeFile(file: vcdPath, text: env.GOVCD_CONFIG_CONTENTS)
        environmentArray << "GOVCD_CONFIG=${vcdPath}"
        temporaryFiles << vcdPath
    } else if (govcdConfigCredentialsID.toLowerCase() != "") {
        // Ensure the path to a VCD parameters file is loaded into the appropriate
        // environment variable for testing scripts to use.
        credentialsArray << [
            $class: 'FileBinding', 
            credentialsId: govcdConfigCredentialsID,
            variable: 'GOVCD_CONFIG'
        ]
    }
}

def build() {
    withCredentials(credentialsArray) {
        withEnv(environmentArray) {
            sh "support/run_in_docker.sh support/build.sh"
        }
    }
}

def cleanupWorkspace() {
    // Remove temporary files
    temporaryFiles.each {
        println "Remove ${it}"
        sh "if [ -f ${it} ]; then rm ${it}; fi"
    }
}

// Call the init method to ensure the environment and credentials are ready. 
init()

// Return a reference to this file to allow the pipeline to call methods. 
return this
