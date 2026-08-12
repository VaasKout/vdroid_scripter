plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.compose)

    alias(libs.plugins.ksp)
    alias(libs.plugins.hilt.android)
    alias(libs.plugins.kotlin.serialization)
}

android {
    namespace = "com.vision.scripter"
    compileSdk {
        version = release(37)
    }

    defaultConfig {
        applicationId = "com.vision.scripter"
        minSdk = 24
        targetSdk = 37
        versionCode = 4
        versionName = "beta-1.0"
        multiDexEnabled = true

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"
    }


    buildTypes {
        release {
            isMinifyEnabled = true
            isShrinkResources = true
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
        }
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
    buildFeatures {
        compose = true
    }
    packaging {
        resources {
            excludes += "/META-INF/{AL2.0,LGPL2.1}"
        }
    }
}

kotlin {
    compilerOptions {
        languageVersion = org.jetbrains.kotlin.gradle.dsl.KotlinVersion.KOTLIN_2_0
        // Optional: Set jvmTarget
        // jvmTarget = org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_11
    }
}

dependencies {

    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.lifecycle.runtime.ktx)
    implementation(libs.androidx.activity.compose)
    implementation(platform(libs.androidx.compose.bom))
    implementation(libs.androidx.ui)
    implementation(libs.androidx.ui.graphics)
    implementation(libs.androidx.ui.tooling.preview)
    implementation(libs.androidx.navigation.compose)
    implementation(libs.androidx.material3)
    implementation(libs.androidx.appcompat.resources)
    implementation(libs.androidx.appcompat)
    testImplementation(libs.junit)
    androidTestImplementation(libs.androidx.junit)
    androidTestImplementation(libs.androidx.espresso.core)
    androidTestImplementation(platform(libs.androidx.compose.bom))
    androidTestImplementation(libs.androidx.ui.test.junit4)
    debugImplementation(libs.androidx.ui.tooling)
    debugImplementation(libs.androidx.ui.test.manifest)
    implementation(libs.kotlinx.serialization.json)
    implementation(libs.androidx.window)
    implementation(libs.androidx.foundation)

    implementation(libs.hilt.android)
    ksp(libs.hilt.android.compiler)
    implementation(libs.androidx.hilt.navigation.compose)

    implementation(libs.kotlinx.coroutines.core)
    implementation(libs.kotlinx.coroutines.android)

    implementation(libs.coil3.coil.compose)
    implementation(libs.coil.network.okhttp)

    implementation(project(":data:api"))
    implementation(project(":data:impl"))
    implementation(project(":feature_welcome:api"))
    implementation(project(":feature_welcome:impl"))
    implementation(project(":feature_main:api"))
    implementation(project(":feature_main:impl"))
    implementation(project(":feature_streaming:api"))
    implementation(project(":feature_streaming:impl"))
    implementation(project(":feature_edit_script:api"))
    implementation(project(":feature_edit_script:impl"))
    implementation(project(":core:ui"))
    implementation(project(":core:coroutines:api"))
    implementation(project(":core:coroutines:impl"))
    implementation(project(":core:prefs:api"))
    implementation(project(":core:prefs:impl"))
    implementation(project(":core:network:api"))
    implementation(project(":core:network:impl"))
}