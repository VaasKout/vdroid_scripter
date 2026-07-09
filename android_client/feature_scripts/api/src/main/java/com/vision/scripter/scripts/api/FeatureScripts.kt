package com.vision.scripter.scripts.api

import androidx.navigation.NavController
import androidx.navigation.NavGraphBuilder

const val ScriptsRoute = "scripts"

interface FeatureScripts {

    fun register(
        navGraphBuilder: NavGraphBuilder,
        navController: NavController,
    )
}