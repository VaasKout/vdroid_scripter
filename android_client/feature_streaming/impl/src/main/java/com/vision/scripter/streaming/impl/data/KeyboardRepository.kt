package com.vision.scripter.streaming.impl.data

import android.view.MotionEvent
import android.view.MotionEvent.ACTION_DOWN
import com.vision.scripter.data.api.ScripterDataSource
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.data.api.models.ScreenSizes
import com.vision.scripter.data.api.models.adjustToClient
import com.vision.scripter.data.api.models.contains
import com.vision.scripter.network.api.ApiResponse
import com.vision.scripter.streaming.impl.di.StreamingScope
import com.vision.scripter.streaming.impl.screen.state.KeyboardMode
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.update
import javax.inject.Inject


@StreamingScope
class KeyboardRepository @Inject constructor(
    private val scripterDataSource: ScripterDataSource,
    private val cvRepository: CvStreamerRepository,
) {

    private val _stateFlow = MutableStateFlow(Keyboard())
    fun observeKeyboardButtons(): Flow<List<RectangleWithText>> = _stateFlow.map { it.buttons }

    private val _selectedButtonFlow = MutableSharedFlow<String>(replay = 1)
    fun observeSelectedButton(): SharedFlow<String> = _selectedButtonFlow.asSharedFlow()

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

    fun handleTouchEvent(event: MotionEvent): String? {
        val state = currentState
        if (state.buttons.isEmpty()) return null

        val newButton = when (state.mode) {
            KeyboardMode.EDIT -> selectKeyboardKey(event)
            KeyboardMode.ADD_NEW -> selectNewKeyboardRect(event)
        }
        if (newButton != null) _selectedButtonFlow.tryEmit(newButton)
        return newButton
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

    fun selectKeyboardKey(event: MotionEvent): String? {
        if (event.action != ACTION_DOWN) return null
        val button = currentState.buttons.firstOrNull {
            it.contains(x = event.x.toInt(), y = event.y.toInt())
        } ?: return null
        cvRepository.setSelectedRectangle(button.rectangle)
        return button.text
    }

    fun selectNewKeyboardRect(event: MotionEvent): String? {
        if (event.action != ACTION_DOWN) return null
        cvRepository.selectRectangle(x = event.x.toInt(), y = event.y.toInt())
        return ""
    }

    fun clear() {
        _stateFlow.update {
            it.copy(
                locale = "",
                buttons = listOf(),
                mode = KeyboardMode.EDIT,
            )
        }
    }
}

data class Keyboard(
    val serial: String = "",
    val screenSizes: ScreenSizes? = null,
    val locale: String = "",
    val buttons: List<RectangleWithText> = listOf(),
    val mode: KeyboardMode = KeyboardMode.EDIT,
)