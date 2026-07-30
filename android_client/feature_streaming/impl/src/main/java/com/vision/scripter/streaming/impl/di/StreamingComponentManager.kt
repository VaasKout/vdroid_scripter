package com.vision.scripter.streaming.impl.di

import javax.inject.Inject
import javax.inject.Provider
import javax.inject.Singleton

@Singleton
class StreamingComponentManager @Inject constructor(
    private val componentBuilder: Provider<StreamingComponentBuilder>,
) {

    @Volatile
    private var component: StreamingComponent? = null

    fun getComponent(): StreamingComponent {
        val current = component
        if (current != null) return current
        return synchronized(this) {
            component ?: componentBuilder.get().build().also { component = it }
        }
    }

    fun destroyComponent() {
        synchronized(this) {
            component = null
        }
    }
}
