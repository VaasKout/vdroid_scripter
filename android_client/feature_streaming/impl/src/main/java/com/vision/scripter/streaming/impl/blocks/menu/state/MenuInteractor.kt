package com.vision.scripter.streaming.impl.blocks.menu.state

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiCommand
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiState
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiStateHolder
import com.vision.scripter.streaming.impl.screen.main.state.CVMode
import com.vision.scripter.streaming.impl.screen.main.state.KeyboardState
import com.vision.scripter.streaming.impl.screen.main.state.increment
import com.vision.scripter.streaming.impl.screen.main.state.toggleDetection
import com.vision.scripter.streaming.impl.shared.MenuToVideo
import com.vision.scripter.streaming.impl.shared.StreamingSharedEventsHolder
import com.vision.scripter.streaming.impl.shared.VideoToMenu
import com.vision.scripter.ui.CommandFlow
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.launchIn
import kotlinx.coroutines.flow.mapNotNull
import kotlinx.coroutines.flow.onEach
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import javax.inject.Inject

@ViewModelScoped
class MenuInteractor @Inject constructor(
    coroutineScopeFactory: CoroutineScopeFactory,
    uiStateMapper: MenuUiStateMapper,
    private val sharedEventsHolder: StreamingSharedEventsHolder,
) : MenuUiStateHolder {

    private val coroutineScope: CoroutineScope =
        coroutineScopeFactory.createBackgroundScope("menu_interactor")

    override val uiCommandsFlow: CommandFlow<MenuUiCommand> =
        CommandFlow(coroutineScope)

    private val _menuState = MutableStateFlow<MenuState>(MenuState.Usual())
    private val menuState = _menuState.asStateFlow()

    private val _dialogState = MutableStateFlow<DialogState>(DialogState.None)
    private val dialogState = _dialogState.asStateFlow()

    override val uiStateFlow: StateFlow<MenuUiState> = combine(
        menuState,
        dialogState,
    ) { menuState, dialogState ->
        uiStateMapper.map(
            state = menuState,
            dialogState = dialogState,
        )
    }.stateIn(coroutineScope, SharingStarted.Eagerly, MenuUiState())

    init {
        sharedEventsHolder.sharedEventsFlow.mapNotNull {
            it as? VideoToMenu
        }.onEach {
            when (it) {
                is VideoToMenu.SelectKeyboardKey -> onEditKeyboardRectangleSelected(it.oldKey)
                is VideoToMenu.SetKeyboardLoading -> setKeyboardLoading(it.isLoading)
            }
        }.launchIn(coroutineScope)
    }

    override fun onScriptModeClicked() {
        _dialogState.update { DialogState.Record }
    }

    override fun onTimeoutClicked() {
        _dialogState.update { DialogState.Timeout }
    }

    override fun onKeyboardClicked() {
        _dialogState.update { DialogState.Keyboard }
    }

    override fun onExpandClicked() {
        val state = _menuState.value
        if (state is MenuState.Usual) {
            _menuState.update { state.copy(expanded = !state.expanded) }
        }
    }

    override fun onRecordingClicked() {
        val state = _menuState.value
        if (state !is MenuState.Recording) return

        _menuState.update { state.copy(controlRecording = !state.controlRecording) }
        sharedEventsHolder.emit(
            MenuToVideo.OnRecordingClicked(isControlRecording = !state.controlRecording)
        )
    }

    override fun onCvModeClicked() {
        when (val state = _menuState.value) {
            is MenuState.Recording -> {
                _menuState.update { MenuState.SelectingCV(localCvMode = CVMode.CV_RECTS) }
                sharedEventsHolder.emit(
                    MenuToVideo.OnCvModeClicked(
                        recording = true,
                        nextMode = CVMode.CV_RECTS,
                    )
                )
            }

            is MenuState.SelectingCV -> {
                val cvMode = state.localCvMode.toggleDetection()
                _menuState.update { state.copy(localCvMode = cvMode) }
                sharedEventsHolder.emit(
                    MenuToVideo.OnCvModeClicked(
                        recording = true,
                        nextMode = cvMode,
                    )
                )
            }

            is MenuState.Usual -> {
                val newCvMode = state.localCvMode.increment()
                _menuState.update { state.copy(localCvMode = newCvMode) }
                sharedEventsHolder.emit(
                    MenuToVideo.OnCvModeClicked(
                        recording = false,
                        nextMode = newCvMode,
                    )
                )
            }

            else -> {}
        }
    }

    override fun onTextModeClicked() {
        val state = _menuState.value
        if (state is MenuState.Usual && state.textHighlighted) {
            _menuState.update { state.copy(textHighlighted = false) }
            sharedEventsHolder.emit(MenuToVideo.OnTextModeClicked(false))
            return
        }
        _dialogState.update { DialogState.Text }
    }

    override fun onTryToFindText(text: String, locale: String) {
        hideDialog()
        sharedEventsHolder.emit(MenuToVideo.FindText(text = text.trim(), locale = locale))
        when (val state = _menuState.value) {
            is MenuState.Recording -> {
                _menuState.update {
                    MenuState.SelectingText
                }
            }

            is MenuState.Usual -> _menuState.update {
                state.copy(localCvMode = CVMode.NO_CV, textHighlighted = true)
            }

            else -> {}
        }
    }

    override fun onKeyboardModeClicked() {
        val state = _menuState.value
        if (state !is MenuState.Keyboard) return

        val newMode = state.mode.increment()
        _menuState.update { state.copy(mode = newMode) }
        sharedEventsHolder.emit(MenuToVideo.KeyboardStateChanged(newMode))
    }

    override fun onEditKeyboardButtonSaved(oldKey: String, newKey: String) {
        val state = _menuState.value
        if (state !is MenuState.Keyboard) return
        hideDialog()
        sharedEventsHolder.emit(MenuToVideo.KeyboardButtonEdited(oldKey, newKey))
    }

    override fun onSaveClicked() {
        sharedEventsHolder.emit(MenuToVideo.SaveClicked)
        when (val state = _menuState.value) {
            is MenuState.Keyboard -> exitKeyboard(state)

            is MenuState.SelectingText, is MenuState.SelectingCV -> {
                _menuState.update { MenuState.Recording() }
            }

            is MenuState.Recording -> {
                _menuState.update { MenuState.Usual(expanded = true) }
            }

            else -> Unit
        }
    }

    override fun onCancelClicked() {
        when (val state = _menuState.value) {
            is MenuState.Recording -> {
                _menuState.update { MenuState.Usual(expanded = true) }
            }

            is MenuState.Keyboard -> exitKeyboard(state)

            else -> {
                _menuState.update { MenuState.Recording() }
            }
        }
        sharedEventsHolder.emit(MenuToVideo.CancelClicked)
    }

    override fun onExitClicked() {
        uiCommandsFlow.tryEmit(MenuUiCommand.ExitCommand)
    }

    override fun onTimeoutSaved(timeout: Int) {
        hideDialog()
        val menuState = _menuState.value
        if (menuState is MenuState.Recording) {
            _menuState.update { menuState.copy(recordTimeout = timeout) }
        }
        sharedEventsHolder.emit(MenuToVideo.TimeoutSaved(timeout))
    }

    override fun onSavedRecordName(name: String) {
        hideDialog()
        _menuState.update { MenuState.Recording() }
        sharedEventsHolder.emit(MenuToVideo.RecordNameSaved(name))
    }

    override fun onSaveLocale(locale: String) {
        hideDialog()
        val updated = when (_menuState.value) {
            is MenuState.Recording -> MenuState.Keyboard(mode = KeyboardState.TYPING)
            is MenuState.Usual -> MenuState.Keyboard(fromUsual = true, mode = KeyboardState.EDIT)
            else -> return
        }
        _menuState.update { updated }
        sharedEventsHolder.emit(
            MenuToVideo.KeyboardLocaleSaved(locale = locale, mode = updated.mode)
        )
    }

    override fun onDialogDismissed() {
        hideDialog()
    }

    private fun setKeyboardLoading(isLoading: Boolean) {
        val state = _menuState.value
        if (state is MenuState.Keyboard) {
            _menuState.value = state.copy(isLoading = isLoading)
        }
    }

    private fun exitKeyboard(state: MenuState.Keyboard) {
        if (state.fromUsual) {
            _menuState.update { MenuState.Usual(expanded = true) }
            return
        }
        _menuState.update { MenuState.Recording() }
    }

    fun onEditKeyboardRectangleSelected(oldKey: String) {
        val state = _menuState.value
        if (state is MenuState.Keyboard && state.mode != KeyboardState.TYPING) {
            _dialogState.update { DialogState.EditKeyboard(oldKey) }
        }
    }

    private fun hideDialog() {
        _dialogState.update { DialogState.None }
    }

    fun clear() {
        _menuState.update { MenuState.Usual() }
        _dialogState.update { DialogState.None }
        coroutineScope.cancel()
    }
}
