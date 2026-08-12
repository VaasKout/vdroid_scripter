package com.vision.scripter

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import com.vision.scripter.main.api.FeatureMain
import com.vision.scripter.streaming.api.FeatureStreaming
import com.vision.scripter.welcome.api.FeatureWelcome
import dagger.Lazy
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class MainActivity : ComponentActivity() {

    @Inject
    lateinit var featureWelcome: Lazy<FeatureWelcome>

    @Inject
    lateinit var featureMain: Lazy<FeatureMain>

    @Inject
    lateinit var featureStreaming: Lazy<FeatureStreaming>

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            AppNavigation(
                featureWelcome = featureWelcome,
                featureMain = featureMain,
                featureStreaming = featureStreaming,
            )
        }
    }
}