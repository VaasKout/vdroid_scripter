package com.vision.scripter.streaming.impl.data

import com.vision.scripter.data.api.ScripterDataSource
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.Parameter
import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.data.api.models.ScreenSizes
import com.vision.scripter.data.api.models.adjustToClient
import com.vision.scripter.data.api.models.contains
import com.vision.scripter.network.api.ApiResponse
import com.vision.scripter.streaming.impl.screen.state.KeyboardMode
import com.vision.scripter.streaming.impl.screen.state.SPACE_KEY
import com.vision.scripter.streaming.impl.screen.state.TYPE_TEXT
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.update
import javax.inject.Inject


@ViewModelScoped
class KeyboardRepository @Inject constructor(
    private val scripterDataSource: ScripterDataSource,
) {

    private val _stateFlow = MutableStateFlow(Keyboard())
    fun observeKeyboardButtons(): Flow<List<RectangleWithText>> = _stateFlow.map { it.buttons }

    private val currentState: Keyboard
        get() = _stateFlow.value

    suspend fun getOrResetKeyboard(
        serial: String,
        screenSizes: ScreenSizes,
    ): Boolean {
        val locale = currentState.locale.ifEmpty { return false }
        val keyboardResult = scripterDataSource.getKeyboard(
            serial = serial,
            locale = locale,
        )
        if (keyboardResult is ApiResponse.Success && keyboardResult.data.isNotEmpty()) {
            setupKeyboardRects(data = keyboardResult.data, screenSizes = screenSizes)
            return true
        }

        val resetResult = scripterDataSource.resetKeyboard(
            serial = serial,
            locale = locale,
        )
        if (resetResult is ApiResponse.Success) {
            setupKeyboardRects(data = resetResult.data, screenSizes = screenSizes)
            return true
        }
        return false
    }

    private fun setupKeyboardRects(
        data: List<RectangleWithText>,
        screenSizes: ScreenSizes,
    ) {
        val buttons = data.mapNotNull {
            val rectangle = it.rectangle ?: return@mapNotNull null
            it.copy(rectangle = rectangle.adjustToClient(screenSizes))
        }

        _stateFlow.update {
            it.copy(buttons = buttons)
        }
    }

    fun updateKeyboardState(keyboardMode: KeyboardMode) {
        _stateFlow.update {
            it.copy(mode = keyboardMode)
        }
    }

    fun updateKeyboardLocale(locale: String) {
        _stateFlow.update {
            it.copy(locale = locale)
        }
    }

    suspend fun editKeyboardKey(
        serial: String,
        oldName: String,
        newName: String,
        rectangle: CvRectangle,
    ): Boolean {
        if (newName.isEmpty()) return false
        currentState.locale.ifEmpty { return false }
        if (oldName.isNotEmpty()) {
            val deleted = scripterDataSource.deleteButton(
                serial = serial,
                locale = currentState.locale,
                name = oldName
            )
            if (!deleted) return false
        }

        return scripterDataSource.editKeyboard(
            serial = serial,
            locale = currentState.locale,
            name = newName,
            rectangle = rectangle,
        )
    }

    fun recordLetters(x: Int, y: Int) {
        val state = currentState
        if (state.locale.isEmpty()) return
        val letter = state.buttons.firstOrNull {
            it.contains(x = x, y = y)
        }?.text ?: return
        if (letter.isEmpty()) return

        val updatedText = buildString {
            append(state.typedText)
            if (letter == SPACE_KEY) append(" ")
            else append(letter)
        }
        _stateFlow.update {
            it.copy(typedText = updatedText)
        }
    }

    fun selectKeyboardKey(x: Int, y: Int): RectangleWithText? {
        return currentState.buttons.firstOrNull {
            it.contains(x = x, y = y)
        }
    }

    fun extractParameter(): Parameter? {
        if (currentState.locale.isEmpty() || currentState.typedText.isEmpty()) return null
        return Parameter(
            type = TYPE_TEXT,
            value = currentState.typedText,
            locale = currentState.locale,
        )
    }

    fun clear() {
        _stateFlow.update {
            it.copy(
                locale = "",
                buttons = listOf(),
                typedText = "",
            )
        }
    }

    fun getMode() = currentState.mode
}

data class Keyboard(
    val serial: String = "",
    val screenSizes: ScreenSizes? = null,
    val locale: String = "",
    val buttons: List<RectangleWithText> = listOf(),
    val mode: KeyboardMode = KeyboardMode.TYPING,
    val typedText: String = "",
)