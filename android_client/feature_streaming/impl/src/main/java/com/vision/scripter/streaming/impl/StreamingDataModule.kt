package com.vision.scripter.streaming.impl

import com.vision.scripter.streaming.impl.data.CvStreamerRepository
import com.vision.scripter.streaming.impl.data.KeyboardRepository
import com.vision.scripter.streaming.impl.data.RecordRepository
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
    fun cvStreamerRepository(): CvStreamerRepository
    fun keyboardRepository(): KeyboardRepository
    fun recordRepository(): RecordRepository
}

@Module
@InstallIn(ViewModelComponent::class)
object StreamingDataModule {

    @Provides
    fun provideVideoStreamerRepository(
        manager: StreamingComponentManager,
    ): VideoStreamerRepository = entryPoint(manager).videoStreamerRepository()

    @Provides
    fun provideCvStreamerRepository(
        manager: StreamingComponentManager,
    ): CvStreamerRepository = entryPoint(manager).cvStreamerRepository()

    @Provides
    fun provideKeyboardRepository(
        manager: StreamingComponentManager,
    ): KeyboardRepository = entryPoint(manager).keyboardRepository()

    @Provides
    fun provideRecordRepository(
        manager: StreamingComponentManager,
    ): RecordRepository = entryPoint(manager).recordRepository()

    private fun entryPoint(manager: StreamingComponentManager): StreamingDataEntryPoint =
        EntryPoints.get(manager.getComponent(), StreamingDataEntryPoint::class.java)
}
