package com.vision.scripter.library.state

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.ScripterDataSource
import com.vision.scripter.network.api.ApiResponse
import com.vision.scripter.ui.CommandFlow
import com.vision.scripter.ui.states.LoadingState
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.cancel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.shareIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject
import kotlin.time.Duration.Companion.milliseconds

@ViewModelScoped
internal class LibraryInteractor @Inject constructor(
    coroutineScopeFactory: CoroutineScopeFactory,
    private val scripterDataSource: ScripterDataSource,
    private val uiStateMapper: LibraryUiStateMapper,
) : LibraryUiStateHolder {

    private val _stateFlow = MutableStateFlow(LibraryState())
    private val stateFlow: StateFlow<LibraryState> = _stateFlow.asStateFlow()

    private val coroutineScope: CoroutineScope =
        coroutineScopeFactory.createBackgroundScope("library_interactor")

    override val uiStateFlow: SharedFlow<LibraryUiState> = stateFlow
        .map(uiStateMapper::map)
        .shareIn(coroutineScope, SharingStarted.WhileSubscribed(), replay = 1)

    override val uiCommandsFlow: CommandFlow<LibraryUiCommand> = CommandFlow(coroutineScope)

    override fun onLoadData(onStart: Boolean) {
        coroutineScope.launch {
            _stateFlow.update {
                it.copy(
                    loadingState = if (onStart) LoadingState.LoadingOnStart
                    else LoadingState.RefreshLoading,
                )
            }

            val libraryLoaded = loadLibrary()
            val routesLoaded = loadRoutes()
            if (!libraryLoaded || !routesLoaded) {
                uiCommandsFlow.tryEmit(LibraryUiCommand.ShowNetworkError)
            }

            if (!onStart) delay(500.milliseconds)
            _stateFlow.update {
                it.copy(loadingState = LoadingState.None)
            }
        }
    }

    override fun onCardClicked(type: LibraryType) {
        uiCommandsFlow.tryEmit(LibraryUiCommand.OpenList(type))
    }

    private suspend fun loadLibrary(): Boolean {
        val result = scripterDataSource.getLibrary()
        if (result !is ApiResponse.Success) return false
        _stateFlow.update {
            it.copy(
                images = result.data.images,
                actions = result.data.actions,
            )
        }
        return true
    }

    private suspend fun loadRoutes(): Boolean {
        val result = scripterDataSource.getRoutes()
        if (result !is ApiResponse.Success) return false
        _stateFlow.update {
            it.copy(routes = result.data)
        }
        return true
    }

    fun clear() {
        coroutineScope.cancel()
    }
}
