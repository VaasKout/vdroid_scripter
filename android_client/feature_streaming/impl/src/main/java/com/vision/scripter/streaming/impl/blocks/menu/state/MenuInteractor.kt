package com.vision.scripter.streaming.impl.blocks.menu.state

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiCommand
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiState
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiStateHolder
import com.vision.scripter.streaming.impl.screen.commandobservers.MenuToVideo
import com.vision.scripter.streaming.impl.screen.commandobservers.VideoToMenu
import com.vision.scripter.streaming.impl.screen.state.CVMode
import com.vision.scripter.streaming.impl.screen.state.KeyboardMode
import com.vision.scripter.streaming.impl.screen.state.increment
import com.vision.scripter.streaming.impl.screen.state.toggleDetection
import com.vision.scripter.ui.CommandFlow
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import javax.inject.Inject

@ViewModelScoped
class MenuInteractor @Inject constructor(
    coroutineScopeFactory: CoroutineScopeFactory,
    uiStateMapper: MenuUiStateMapper,
) : MenuUiStateHolder {

    private val coroutineScope: CoroutineScope =
        coroutineScopeFactory.createBackgroundScope("menu_interactor")

    override val uiCommandsFlow: CommandFlow<MenuUiCommand> = CommandFlow(coroutineScope)
    override val videoCommandsFlow: CommandFlow<MenuToVideo> = CommandFlow(coroutineScope)

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

    override fun onSharedEvent(event: VideoToMenu) {
        when (event) {
            is VideoToMenu.SelectKeyboardKey -> onEditKeyboardRectangleSelected(event.oldKey)
            is VideoToMenu.SetKeyboardLoading -> setKeyboardLoading(event.isLoading)
        }
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
        videoCommandsFlow.tryEmit(
            MenuToVideo.OnRecordingClicked(isControlRecording = !state.controlRecording)
        )
    }

    override fun onCvModeClicked() {
        when (val state = _menuState.value) {
            is MenuState.Recording -> {
                _menuState.update { MenuState.SelectingCV(localCvMode = CVMode.CV_RECTS) }
                videoCommandsFlow.tryEmit(
                    MenuToVideo.OnCvModeClicked(
                        recording = true,
                        nextMode = CVMode.CV_RECTS,
                    )
                )
            }

            is MenuState.SelectingCV -> {
                val cvMode = state.localCvMode.toggleDetection()
                _menuState.update { state.copy(localCvMode = cvMode) }
                videoCommandsFlow.tryEmit(
                    MenuToVideo.OnCvModeClicked(
                        recording = true,
                        nextMode = cvMode,
                    )
                )
            }

            is MenuState.Usual -> {
                val newCvMode = state.localCvMode.increment()
                _menuState.update { state.copy(localCvMode = newCvMode) }
                videoCommandsFlow.tryEmit(
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
            videoCommandsFlow.tryEmit(MenuToVideo.OnTextModeClicked(false))
            return
        }
        _dialogState.update { DialogState.Text }
    }

    override fun onTryToFindText(text: String, locale: String) {
        hideDialog()
        videoCommandsFlow.tryEmit(MenuToVideo.FindText(text = text.trim(), locale = locale))
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
        videoCommandsFlow.tryEmit(MenuToVideo.KeyboardStateChanged(newMode))
    }

    override fun onEditKeyboardButtonSaved(oldKey: String, newKey: String) {
        val state = _menuState.value
        if (state !is MenuState.Keyboard) return
        hideDialog()
        videoCommandsFlow.tryEmit(MenuToVideo.KeyboardButtonEdited(oldKey, newKey))
    }

    override fun onSaveClicked() {
        videoCommandsFlow.tryEmit(MenuToVideo.SaveClicked)
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
        videoCommandsFlow.tryEmit(MenuToVideo.CancelClicked)
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
        videoCommandsFlow.tryEmit(MenuToVideo.TimeoutSaved(timeout))
    }

    override fun onSavedRecordName(name: String) {
        hideDialog()
        _menuState.update { MenuState.Recording() }
        videoCommandsFlow.tryEmit(MenuToVideo.RecordNameSaved(name))
    }

    override fun onSaveLocale(locale: String) {
        hideDialog()
        val updated = when (_menuState.value) {
            is MenuState.Recording -> MenuState.Keyboard(mode = KeyboardMode.TYPING)
            is MenuState.Usual -> MenuState.Keyboard(fromUsual = true, mode = KeyboardMode.TYPING)
            else -> return
        }
        _menuState.update { updated }
        videoCommandsFlow.tryEmit(MenuToVideo.KeyboardLocaleSaved(locale = locale))
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

    private fun onEditKeyboardRectangleSelected(oldKey: String) {
        val state = _menuState.value
        if (state is MenuState.Keyboard && state.mode != KeyboardMode.TYPING) {
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
