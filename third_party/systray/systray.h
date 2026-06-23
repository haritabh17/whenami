#include "stdbool.h"

extern void systray_ready();
extern void systray_on_exit();
extern void systray_menu_item_selected(int menu_id);
extern void systray_menu_item_slider_changed(int menu_id, int value);
void registerSystray(void);
int nativeLoop(void);

void setIcon(const char* iconBytes, int length, bool template);
void setMenuItemIcon(const char* iconBytes, int length, int menuId, bool template, int avatarSize, int displaySize);
void setTitle(char* title);
void setTooltip(char* tooltip);

typedef struct {
  char* text;
  char* image_bytes;
  int image_len;
  int avatar_size;
  int display_size;
} status_segment_t;

void setStatusSegments(status_segment_t* segments, int count);
void add_or_update_menu_item(int menuId, int parentMenuId, char* title, char* tooltip, short disabled, short checked, short isCheckable);
void set_menu_item_slider(int menuId, int minValue, int maxValue, int value);
void add_separator(int menuId);
void hide_menu_item(int menuId);
void show_menu_item(int menuId);
void quit();
