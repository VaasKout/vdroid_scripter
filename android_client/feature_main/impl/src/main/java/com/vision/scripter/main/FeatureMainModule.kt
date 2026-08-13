package com.vision.scripter.main

import com.vision.scripter.main.api.FeatureMain
import dagger.Binds
import dagger.Module
import dagger.hilt.InstallIn
import dagger.hilt.android.components.ActivityComponent

@Module
@InstallIn(ActivityComponent::class)
interface FeatureMainModule {

    @Binds
    fun bindMainScreen(featureMainImpl: FeatureMainImpl): FeatureMain
}