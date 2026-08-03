package com.vision.scripter

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.rememberNavController
import com.vision.scripter.editscript.api.FeatureEditScript
import com.vision.scripter.main.api.FeatureMain
import com.vision.scripter.streaming.api.FeatureStreaming
import com.vision.scripter.welcome.api.FeatureWelcome
import com.vision.scripter.welcome.api.WelcomeRoute
import dagger.Lazy

@Composable
fun AppNavigation(
    featureWelcome: Lazy<FeatureWelcome>,
    featureMain: Lazy<FeatureMain>,
    featureStreaming: Lazy<FeatureStreaming>,
    featureEditScript: Lazy<FeatureEditScript>,
) {
    val navController = rememberNavController()
    NavHost(
        modifier = Modifier
            .fillMaxSize()
            .background(color = MaterialTheme.colorScheme.background),
        startDestination = WelcomeRoute,
        navController = navController,
    ) {
        featureWelcome.get().register(
            navGraphBuilder = this,
            navController = navController,
        )
        featureMain.get().register(
            navGraphBuilder = this,
            navController = navController,
        )
        featureStreaming.get().register(
            navGraphBuilder = this,
            navController = navController,
        )
        featureEditScript.get().register(
            navGraphBuilder = this,
            navController = navController,
        )
    }
}
