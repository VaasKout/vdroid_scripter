package com.vision.scripter.welcome.impl.state

import androidx.core.net.toUri
import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.ScripterRepository
import com.vision.scripter.prefs.api.DataStoreRepository
import com.vision.scripter.ui.CommandFlow
import com.vision.scripter.welcome.impl.ui.WelcomeUiCommand
import com.vision.scripter.welcome.impl.ui.WelcomeUiState
import com.vision.scripter.welcome.impl.ui.WelcomeUiStateHolder
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@ViewModelScoped
class WelcomeInteractor @Inject constructor(
    coroutineScopeFactory: CoroutineScopeFactory,
    private val uiStateMapper: WelcomeStateUiMapper,
    private val dataStoreRepository: DataStoreRepository,
    private val scripterRepository: ScripterRepository,
) : WelcomeUiStateHolder {

    private val _stateFlow = MutableStateFlow(WelcomeState())
    private val stateFlow: StateFlow<WelcomeState> = _stateFlow.asStateFlow()
    private val coroutineScope: CoroutineScope =
        coroutineScopeFactory.createBackgroundScope("welcome_interactor")

    override val uiStateFlow: StateFlow<WelcomeUiState?>
        get() = stateFlow.map(uiStateMapper::map)
            .stateIn(coroutineScope, SharingStarted.WhileSubscribed(), initialValue = null)

    override val uiCommandsFlow: CommandFlow<WelcomeUiCommand> = CommandFlow(coroutineScope)

    override fun onInitData() {
        coroutineScope.launch {
            val fullUrl = dataStoreRepository.getServerUrl()
            if (fullUrl.isEmpty()) return@launch

            val uri = fullUrl.toUri()
            val port = uri.port
            val url = "${uri.scheme}://${uri.host}"
            _stateFlow.update {
                it.copy(
                    oldUrl = url,
                    oldPort = if (port <= 0) "" else port.toString(),
                )
            }
            val serverAvailable = scripterRepository.pingServer()
            if (serverAvailable) {
                uiCommandsFlow.tryEmit(WelcomeUiCommand.NavigateToMain)
            }
        }
    }

    override fun onApplyData(url: String, port: String) {
        coroutineScope.launch {
            if (url.isEmpty()) {
                uiCommandsFlow.tryEmit(WelcomeUiCommand.ShowAddressError)
                return@launch
            }
            val uri = url.toUri()

            val fullUrl = when {
                uri.scheme != null && uri.host != null -> {
                    buildString {
                        append(uri.scheme)
                        append("://")
                        append(uri.host)
                        if (port.isNotEmpty()) append(":$port")
                    }
                }

                uri.path != null -> "http://${uri.path}:${port}"
                else -> ""
            }

            if (fullUrl.isEmpty()) {
                uiCommandsFlow.tryEmit(WelcomeUiCommand.ShowAddressError)
                return@launch
            }

            dataStoreRepository.setServerUrl(fullUrl)
            val serverAvailable = scripterRepository.pingServer()
            if (!serverAvailable) {
                uiCommandsFlow.tryEmit(WelcomeUiCommand.ShowAddressError)
                return@launch
            }
            uiCommandsFlow.tryEmit(WelcomeUiCommand.NavigateToMain)
        }
    }

    fun clear() {
        coroutineScope.cancel()
    }
}
