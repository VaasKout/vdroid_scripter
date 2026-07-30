package com.vision.scripter.streaming.impl.di

import dagger.hilt.DefineComponent
import dagger.hilt.components.SingletonComponent

@StreamingScope
@DefineComponent(parent = SingletonComponent::class)
interface StreamingComponent

@DefineComponent.Builder
interface StreamingComponentBuilder {
    fun build(): StreamingComponent
}
