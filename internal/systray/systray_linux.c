#include "systray.h"

// Systray is disabled on Linux for this application.
// Provide no-op implementations to satisfy the linker.

void registerSystray(void) {
    systray_ready();
}

int nativeLoop(void) {
    return 0;
}

void setIcon(const char *iconBytes, int length, bool template) {}
void setMenuItemIcon(const char *iconBytes, int length, int menuId, bool template) {}
void setTitle(char *title) { if (title) free(title); }
void setTooltip(char *tooltip) { if (tooltip) free(tooltip); }
void add_or_update_menu_item(int menuId, int parentMenuId, char *title, char *tooltip, short disabled, short checked, short isCheckable) {
    if (title) free(title);
    if (tooltip) free(tooltip);
}
void add_separator(int menuId) {}
void hide_menu_item(int menuId) {}
void show_menu_item(int menuId) {}
void stop_systray(void) {}
void forceQuit(void) {}
