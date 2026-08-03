package com.vision.scripter.editscript.impl

import com.vision.scripter.editscript.api.FeatureEditScript
import dagger.Binds
import dagger.Module
import dagger.hilt.InstallIn
import dagger.hilt.android.components.ActivityComponent

@Module
@InstallIn(ActivityComponent::class)
interface FeatureEditScriptModule {

    @Binds
    fun bindEditScriptScreen(featureEditScriptImpl: FeatureEditScriptImpl): FeatureEditScript
}
