#import <Cocoa/Cocoa.h>
#include <math.h>
#include "systray.h"

#if __MAC_OS_X_VERSION_MIN_REQUIRED < 101400

    #ifndef NSControlStateValueOff
      #define NSControlStateValueOff NSOffState
    #endif

    #ifndef NSControlStateValueOn
      #define NSControlStateValueOn NSOnState
    #endif

#endif

static NSImage* scaledIconFromData(NSData *buffer, CGFloat size, bool template);
static NSImage* scaledAvatarIconFromData(NSData *buffer, CGFloat displaySize, int avatarPx, bool template);

static const NSInteger avatarMenuImageTag = 1001;
static const NSInteger avatarMenuTextTag = 1002;

@interface MenuItem : NSObject
{
  @public
    NSNumber* menuId;
    NSNumber* parentMenuId;
    NSString* title;
    NSString* tooltip;
    short disabled;
    short checked;
}
-(id) initWithId: (int)theMenuId
withParentMenuId: (int)theParentMenuId
       withTitle: (const char*)theTitle
     withTooltip: (const char*)theTooltip
    withDisabled: (short)theDisabled
     withChecked: (short)theChecked;
     @end
     @implementation MenuItem
     -(id) initWithId: (int)theMenuId
     withParentMenuId: (int)theParentMenuId
            withTitle: (const char*)theTitle
          withTooltip: (const char*)theTooltip
         withDisabled: (short)theDisabled
          withChecked: (short)theChecked
{
  menuId = [NSNumber numberWithInt:theMenuId];
  parentMenuId = [NSNumber numberWithInt:theParentMenuId];
  title = [[NSString alloc] initWithCString:theTitle
                                   encoding:NSUTF8StringEncoding];
  tooltip = [[NSString alloc] initWithCString:theTooltip
                                     encoding:NSUTF8StringEncoding];
  disabled = theDisabled;
  checked = theChecked;
  return self;
}
@end

@interface AppDelegate: NSObject <NSApplicationDelegate>
  - (void) add_or_update_menu_item:(MenuItem*) item;
  - (void) refreshAvatarMenuItemView:(NSMenuItem*) menuItem;
  - (IBAction)menuHandler:(id)sender;
  @property (assign) IBOutlet NSWindow *window;
  @end

  @implementation AppDelegate
{
  NSStatusItem *statusItem;
  NSMenu *menu;
  NSCondition* cond;
}

@synthesize window = _window;

- (void)applicationDidFinishLaunching:(NSNotification *)aNotification
{
  self->statusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
  self->menu = [[NSMenu alloc] init];
  [self->menu setAutoenablesItems: FALSE];
  [self->statusItem setMenu:self->menu];
  systray_ready();
}

- (void)applicationWillTerminate:(NSNotification *)aNotification
{
  systray_on_exit();
}

- (void)setIcon:(NSImage *)image {
  statusItem.button.image = image;
  [self updateTitleButtonStyle];
}

- (void)setTitle:(NSString *)title {
  statusItem.button.attributedTitle = [[NSAttributedString alloc] init];
  statusItem.button.title = title;
  [self updateTitleButtonStyle];
}

- (void)setStatusSegments:(NSArray *)segments {
  NSFont *font = [NSFont menuBarFontOfSize:0];
  if (font == nil) {
    font = [NSFont systemFontOfSize:13];
  }
  NSMutableAttributedString *result = [[NSMutableAttributedString alloc] init];
  NSDictionary *textAttrs = @{NSFontAttributeName: font};
  const CGFloat defaultImgSize = 30.0;

  for (NSDictionary *seg in segments) {
    NSData *imgData = seg[@"image"];
    NSString *text = seg[@"text"];

    if (imgData != nil && [imgData length] > 0) {
      int avatarPx = 0;
      if (seg[@"avatarSize"] != nil) {
        avatarPx = [seg[@"avatarSize"] intValue];
      }
      CGFloat imgSize = defaultImgSize;
      if (seg[@"displaySize"] != nil && [seg[@"displaySize"] intValue] > 0) {
        imgSize = (CGFloat)[seg[@"displaySize"] intValue];
      }
      CGFloat imgBaseline = round((font.capHeight - imgSize) / 2.0);
      NSImage *img = scaledAvatarIconFromData(imgData, imgSize, avatarPx, false);
      if (img != nil) {
        NSTextAttachment *attachment = [[NSTextAttachment alloc] init];
        attachment.image = img;
        CGFloat boundH = imgSize;
        CGFloat boundW = imgSize;
        if (img.size.height > 0 && img.size.width > 0) {
          boundW = boundH * (img.size.width / img.size.height);
        }
        attachment.bounds = NSMakeRect(0, imgBaseline, boundW, boundH);
        NSAttributedString *imgStr = [NSAttributedString attributedStringWithAttachment:attachment];
        [result appendAttributedString:imgStr];
        [result appendAttributedString:[[NSAttributedString alloc] initWithString:@" " attributes:textAttrs]];
      }
    }

    if (text != nil && [text length] > 0) {
      [result appendAttributedString:[[NSAttributedString alloc] initWithString:text attributes:textAttrs]];
    }
  }

  statusItem.button.image = nil;
  statusItem.button.title = @"";
  statusItem.button.attributedTitle = result;
  statusItem.button.imagePosition = NSNoImage;
}

-(void)updateTitleButtonStyle {
  if (statusItem.button.image != nil) {
    if ([statusItem.button.title length] == 0) {
      statusItem.button.imagePosition = NSImageOnly;
    } else {
      statusItem.button.imagePosition = NSImageLeft;
    }
  } else {
    statusItem.button.imagePosition = NSNoImage;
  }
}


- (void)setTooltip:(NSString *)tooltip {
  statusItem.button.toolTip = tooltip;
}

- (IBAction)menuHandler:(id)sender
{
  NSNumber* menuId = [sender representedObject];
  systray_menu_item_selected(menuId.intValue);
}

- (IBAction)sliderHandler:(id)sender
{
  NSSlider *slider = (NSSlider *)sender;
  systray_menu_item_slider_changed((int)slider.tag, (int)lround(slider.doubleValue));
}

- (void)add_or_update_menu_item:(MenuItem *)item {
  NSMenu *theMenu = self->menu;
  NSMenuItem *parentItem;
  if ([item->parentMenuId integerValue] > 0) {
    parentItem = find_menu_item(menu, item->parentMenuId);
    if (parentItem.hasSubmenu) {
      theMenu = parentItem.submenu;
    } else {
      theMenu = [[NSMenu alloc] init];
      [theMenu setAutoenablesItems:NO];
      [parentItem setSubmenu:theMenu];
    }
  }
  
  NSMenuItem *menuItem;
  menuItem = find_menu_item(theMenu, item->menuId);
  if (menuItem == NULL) {
    menuItem = [theMenu addItemWithTitle:item->title
                               action:@selector(menuHandler:)
                        keyEquivalent:@""];
    [menuItem setRepresentedObject:item->menuId];
  }
  [menuItem setTitle:item->title];
  [menuItem setTag:[item->menuId integerValue]];
  [menuItem setTarget:self];
  [menuItem setToolTip:item->tooltip];
  if (item->disabled == 1) {
    menuItem.enabled = FALSE;
  } else {
    menuItem.enabled = TRUE;
  }
  if (item->checked == 1) {
    menuItem.state = NSControlStateValueOn;
  } else {
    menuItem.state = NSControlStateValueOff;
  }
  [self refreshAvatarMenuItemView:menuItem];
}

- (void) refreshAvatarMenuItemView:(NSMenuItem*) menuItem {
  NSView *view = menuItem.view;
  NSImageView *imageView = [view viewWithTag:avatarMenuImageTag];
  if (imageView == nil || imageView.image == nil) {
    return;
  }
  [self setAvatarMenuItemView:menuItem image:imageView.image displaySize:imageView.frame.size.height];
}

- (void) setAvatarMenuItemView:(NSMenuItem*) menuItem image:(NSImage*) image displaySize:(CGFloat)displaySize {
  if (displaySize <= 0) {
    displaySize = 18.0;
  }

  NSFont *font = [NSFont menuFontOfSize:0];
  if (font == nil) {
    font = [NSFont systemFontOfSize:13];
  }

  NSView *view = menuItem.view;
  NSImageView *imageView = [view viewWithTag:avatarMenuImageTag];
  NSTextField *textField = [view viewWithTag:avatarMenuTextTag];
  if (imageView == nil || textField == nil) {
    view = [[NSView alloc] initWithFrame:NSMakeRect(0, 0, 1, 1)];

    imageView = [[NSImageView alloc] initWithFrame:NSZeroRect];
    imageView.tag = avatarMenuImageTag;
    imageView.imageScaling = NSImageScaleProportionallyUpOrDown;
    [view addSubview:imageView];

    textField = [[NSTextField alloc] initWithFrame:NSZeroRect];
    textField.tag = avatarMenuTextTag;
    textField.editable = NO;
    textField.selectable = NO;
    textField.bezeled = NO;
    textField.drawsBackground = NO;
    [view addSubview:textField];
  }

  NSString *title = menuItem.title != nil ? menuItem.title : @"";
  CGFloat imageW = displaySize;
  if (image.size.height > 0 && image.size.width > 0) {
    imageW = displaySize * (image.size.width / image.size.height);
  }
  CGFloat textH = ceil(font.ascender - font.descender + font.leading);
  CGFloat textW = ceil([title sizeWithAttributes:@{NSFontAttributeName: font}].width);
  CGFloat height = ceil(MAX(24.0, MAX(displaySize + 6.0, textH + 8.0)));
  CGFloat imageX = 12.0;
  CGFloat textX = imageX + imageW + 8.0;
  CGFloat width = ceil(textX + textW + 24.0);

  view.frame = NSMakeRect(0, 0, width, height);
  imageView.frame = NSMakeRect(imageX, round((height - displaySize) / 2.0), imageW, displaySize);
  imageView.image = image;

  textField.font = font;
  textField.stringValue = title;
  textField.textColor = menuItem.enabled ? [NSColor textColor] : [NSColor disabledControlTextColor];
  textField.frame = NSMakeRect(textX, round((height - textH) / 2.0), textW + 4.0, textH);

  menuItem.view = view;
  [view setNeedsDisplay:YES];
  [imageView setNeedsDisplay:YES];
  [textField setNeedsDisplay:YES];
  [menuItem.menu update];
}

- (void) setMenuItemSlider:(NSArray*)sliderConfig {
  NSNumber* menuId = [sliderConfig objectAtIndex:0];
  NSNumber* minValue = [sliderConfig objectAtIndex:1];
  NSNumber* maxValue = [sliderConfig objectAtIndex:2];
  NSNumber* value = [sliderConfig objectAtIndex:3];

  NSMenuItem* menuItem = find_menu_item(menu, menuId);
  if (menuItem == NULL) {
    return;
  }

  NSSlider *slider = [[NSSlider alloc] initWithFrame:NSMakeRect(12, 0, 180, 24)];
  slider.minValue = minValue.doubleValue;
  slider.maxValue = maxValue.doubleValue;
  slider.doubleValue = value.doubleValue;
  slider.continuous = YES;
  slider.target = self;
  slider.action = @selector(sliderHandler:);
  slider.tag = menuId.integerValue;

  NSView *view = [[NSView alloc] initWithFrame:NSMakeRect(0, 0, 204, 24)];
  [view addSubview:slider];
  menuItem.view = view;
}

NSMenuItem *find_menu_item(NSMenu *ourMenu, NSNumber *menuId) {
  NSMenuItem *foundItem = [ourMenu itemWithTag:[menuId integerValue]];
  if (foundItem != NULL) {
    return foundItem;
  }
  NSArray *menu_items = ourMenu.itemArray;
  int i;
  for (i = 0; i < [menu_items count]; i++) {
    NSMenuItem *i_item = [menu_items objectAtIndex:i];
    if (i_item.hasSubmenu) {
      foundItem = find_menu_item(i_item.submenu, menuId);
      if (foundItem != NULL) {
        return foundItem;
      }
    }
  }

  return NULL;
};

- (void) add_separator:(NSNumber*) menuId
{
  [menu addItem: [NSMenuItem separatorItem]];
}

- (void) hide_menu_item:(NSNumber*) menuId
{
  NSMenuItem* menuItem = find_menu_item(menu, menuId);
  if (menuItem != NULL) {
    [menuItem setHidden:TRUE];
  }
}

- (void) setMenuItemIcon:(NSArray*)imageAndMenuId {
  NSImage* image = [imageAndMenuId objectAtIndex:0];
  NSNumber* menuId = [imageAndMenuId objectAtIndex:1];
  CGFloat displaySize = 0;
  if ([imageAndMenuId count] > 2) {
    displaySize = [[imageAndMenuId objectAtIndex:2] doubleValue];
  }

  NSMenuItem* menuItem;
  menuItem = find_menu_item(menu, menuId);
  if (menuItem == NULL) {
    return;
  }
  if (displaySize > 0) {
    [self setAvatarMenuItemView:menuItem image:image displaySize:displaySize];
    return;
  }
  image.cacheMode = NSImageCacheNever;
  menuItem.image = nil;
  menuItem.image = image;
  [menuItem.menu update];
}

- (void) show_menu_item:(NSNumber*) menuId
{
  NSMenuItem* menuItem = find_menu_item(menu, menuId);
  if (menuItem != NULL) {
    [menuItem setHidden:FALSE];
  }
}

- (void) quit
{
  [NSApp terminate:self];
}

@end

void registerSystray(void) {
  AppDelegate *delegate = [[AppDelegate alloc] init];
  [[NSApplication sharedApplication] setDelegate:delegate];
  // A workaround to avoid crashing on macOS versions before Catalina. Somehow
  // SIGSEGV would happen inside AppKit if [NSApp run] is called from a
  // different function, even if that function is called right after this.
  if (floor(NSAppKitVersionNumber) <= /*NSAppKitVersionNumber10_14*/ 1671){
    [NSApp run];
  }
}

int nativeLoop(void) {
  if (floor(NSAppKitVersionNumber) > /*NSAppKitVersionNumber10_14*/ 1671){
    [NSApp run];
  }
  return EXIT_SUCCESS;
}

void runInMainThread(SEL method, id object) {
  [(AppDelegate*)[NSApp delegate]
    performSelectorOnMainThread:method
                     withObject:object
                  waitUntilDone: YES];
}

static NSImage* scaledAvatarIconFromData(NSData *buffer, CGFloat displaySize, int avatarPx, bool template) {
  if (buffer == nil || [buffer length] == 0) {
    return nil;
  }
  NSImage *source = [[NSImage alloc] initWithData:buffer];
  if (source == nil) {
    return nil;
  }
  if (avatarPx <= 0) {
    return scaledIconFromData(buffer, displaySize, template);
  }

  CGFloat scale = displaySize / (CGFloat)avatarPx;
  NSSize srcSize = source.size;
  CGFloat outW = srcSize.width * scale;
  CGFloat outH = srcSize.height * scale;

  NSImage *image = [[NSImage alloc] initWithSize:NSMakeSize(outW, outH)];
  [image lockFocus];
  [[NSGraphicsContext currentContext] setImageInterpolation:NSImageInterpolationHigh];
  NSRect dest = NSMakeRect(0, 0, outW, outH);
  NSRect src = NSMakeRect(0, 0, srcSize.width, srcSize.height);
  [source drawInRect:dest fromRect:src operation:NSCompositingOperationSourceOver fraction:1.0 respectFlipped:YES hints:nil];
  [image unlockFocus];
  image.cacheMode = NSImageCacheNever;
  image.template = template;
  return image;
}

static NSImage* scaledIconFromData(NSData *buffer, CGFloat size, bool template) {
  if (buffer == nil || [buffer length] == 0) {
    return nil;
  }
  NSImage *source = [[NSImage alloc] initWithData:buffer];
  if (source == nil) {
    return nil;
  }
  NSImage *image = [[NSImage alloc] initWithSize:NSMakeSize(size, size)];
  [image lockFocus];
  [[NSGraphicsContext currentContext] setImageInterpolation:NSImageInterpolationHigh];
  NSRect dest = NSMakeRect(0, 0, size, size);
  [source drawInRect:dest fromRect:NSZeroRect operation:NSCompositingOperationSourceOver fraction:1.0 respectFlipped:YES hints:nil];
  [image unlockFocus];
  image.cacheMode = NSImageCacheNever;
  image.template = template;
  return image;
}

static NSImage* scaledIconFromBytes(const char* iconBytes, int length, CGFloat size, bool template) {
  return scaledIconFromData([NSData dataWithBytes:iconBytes length:length], size, template);
}

void setIcon(const char* iconBytes, int length, bool template) {
  NSImage *image = scaledIconFromBytes(iconBytes, length, 16.0, template);
  if (image == nil) {
    return;
  }
  runInMainThread(@selector(setIcon:), (id)image);
}

void setMenuItemIcon(const char* iconBytes, int length, int menuId, bool template, int avatarSize, int displaySize) {
  CGFloat imgSize = displaySize > 0 ? (CGFloat)displaySize : 18.0;
  NSImage *image = scaledAvatarIconFromData([NSData dataWithBytes:iconBytes length:length], imgSize, avatarSize, template);
  if (image == nil) {
    return;
  }
  NSNumber *mId = [NSNumber numberWithInt:menuId];
  NSNumber *display = [NSNumber numberWithInt:displaySize];
  runInMainThread(@selector(setMenuItemIcon:), @[image, (id)mId, display]);
}

void set_menu_item_slider(int menuId, int minValue, int maxValue, int value) {
  NSNumber *mId = [NSNumber numberWithInt:menuId];
  NSNumber *min = [NSNumber numberWithInt:minValue];
  NSNumber *max = [NSNumber numberWithInt:maxValue];
  NSNumber *current = [NSNumber numberWithInt:value];
  runInMainThread(@selector(setMenuItemSlider:), @[mId, min, max, current]);
}

void setTitle(char* ctitle) {
  NSString* title = [[NSString alloc] initWithCString:ctitle
                                             encoding:NSUTF8StringEncoding];
  free(ctitle);
  runInMainThread(@selector(setTitle:), (id)title);
}

void setTooltip(char* ctooltip) {
  NSString* tooltip = [[NSString alloc] initWithCString:ctooltip
                                               encoding:NSUTF8StringEncoding];
  free(ctooltip);
  runInMainThread(@selector(setTooltip:), (id)tooltip);
}

void setStatusSegments(status_segment_t* segments, int count) {
  if (segments == NULL || count <= 0) {
    return;
  }
  NSMutableArray *arr = [NSMutableArray arrayWithCapacity:count];
  for (int i = 0; i < count; i++) {
    NSMutableDictionary *d = [NSMutableDictionary dictionary];
    if (segments[i].text != NULL) {
      NSString *text = [[NSString alloc] initWithUTF8String:segments[i].text];
      if (text != nil) {
        d[@"text"] = text;
      }
    }
    if (segments[i].image_len > 0 && segments[i].image_bytes != NULL) {
      d[@"image"] = [NSData dataWithBytes:segments[i].image_bytes length:segments[i].image_len];
    }
    if (segments[i].avatar_size > 0) {
      d[@"avatarSize"] = @(segments[i].avatar_size);
    }
    if (segments[i].display_size > 0) {
      d[@"displaySize"] = @(segments[i].display_size);
    }
    [arr addObject:d];
  }
  runInMainThread(@selector(setStatusSegments:), arr);
}

void add_or_update_menu_item(int menuId, int parentMenuId, char* title, char* tooltip, short disabled, short checked, short isCheckable) {
  MenuItem* item = [[MenuItem alloc] initWithId: menuId withParentMenuId: parentMenuId withTitle: title withTooltip: tooltip withDisabled: disabled withChecked: checked];
  free(title);
  free(tooltip);
  runInMainThread(@selector(add_or_update_menu_item:), (id)item);
}

void add_separator(int menuId) {
  NSNumber *mId = [NSNumber numberWithInt:menuId];
  runInMainThread(@selector(add_separator:), (id)mId);
}

void hide_menu_item(int menuId) {
  NSNumber *mId = [NSNumber numberWithInt:menuId];
  runInMainThread(@selector(hide_menu_item:), (id)mId);
}

void show_menu_item(int menuId) {
  NSNumber *mId = [NSNumber numberWithInt:menuId];
  runInMainThread(@selector(show_menu_item:), (id)mId);
}

void quit() {
  runInMainThread(@selector(quit), nil);
}
