package com.vision.scripter.streaming.impl.di

import com.vision.scripter.streaming.impl.data.CvRepository
import com.vision.scripter.streaming.impl.data.RecordRepository
import com.vision.scripter.streaming.impl.screen.StreamingEventsHolder
import com.vision.scripter.streaming.impl.data.VideoStreamerRepository
import dagger.Module
import dagger.Provides
import dagger.hilt.EntryPoint
import dagger.hilt.EntryPoints
import dagger.hilt.InstallIn
import dagger.hilt.android.components.ViewModelComponent

@EntryPoint
@InstallIn(StreamingComponent::class)
interface StreamingDataEntryPoint {
    fun videoStreamerRepository(): VideoStreamerRepository
    fun cvRepository(): CvRepository
    fun recordRepository(): RecordRepository
    fun streamingEventRepository(): StreamingEventsHolder
}

@Module
@InstallIn(ViewModelComponent::class)
object StreamingDataModule {

    @Provides
    fun provideVideoStreamerRepository(
        manager: StreamingComponentManager,
    ): VideoStreamerRepository = entryPoint(manager).videoStreamerRepository()

    @Provides
    fun provideCvRepository(
        manager: StreamingComponentManager,
    ): CvRepository = entryPoint(manager).cvRepository()

    @Provides
    fun provideRecordRepository(
        manager: StreamingComponentManager,
    ): RecordRepository = entryPoint(manager).recordRepository()

    @Provides
    fun provideStreamingEventRepository(
        manager: StreamingComponentManager,
    ): StreamingEventsHolder = entryPoint(manager).streamingEventRepository()

    private fun entryPoint(manager: StreamingComponentManager): StreamingDataEntryPoint =
        EntryPoints.get(manager.getComponent(), StreamingDataEntryPoint::class.java)
}
