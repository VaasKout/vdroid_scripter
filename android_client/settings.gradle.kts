pluginManagement {
    repositories {
        google {
            content {
                includeGroupByRegex("com\\.android.*")
                includeGroupByRegex("com\\.google.*")
                includeGroupByRegex("androidx.*")
            }
        }
        mavenCentral()
        gradlePluginPortal()
    }
}
plugins {
    id("org.gradle.toolchains.foojay-resolver-convention") version "1.0.0"
}
dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
        google()
        mavenCentral()
    }
}

rootProject.name = "VisionScripter"
include(":app")
include(":data:api")
include(":data:impl")
include(":core:ui")
include(":feature_welcome:api")
include(":feature_welcome:impl")
include(":core:")
include(":feature_main")
include(":feature_main:api")
include(":feature_main:impl")
include(":core:coroutines:api")
include(":core:coroutines:impl")
include(":core:network:api")
include(":core:network:impl")
include(":core:prefs:api")
include(":core:prefs:impl")
include(":feature_streaming:api")
include(":feature_streaming:impl")
