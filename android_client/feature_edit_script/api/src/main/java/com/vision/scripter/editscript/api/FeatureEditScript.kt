package com.vision.scripter.editscript.api

import androidx.navigation.NavController
import androidx.navigation.NavGraphBuilder

const val EditScriptRoute = "edit_script"
const val EditScriptRouteWithArgs = "edit_script/{location}/{name}"
const val EditScriptLocationArg = "location"
const val EditScriptNameArg = "name"

interface FeatureEditScript {

    fun register(
        navGraphBuilder: NavGraphBuilder,
        navController: NavController,
    )
}
