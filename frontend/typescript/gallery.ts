declare var ResizeObserver;


let __UNIQUE_ID__ = 1;
export abstract class GalleryImage {
  private __UID: number;
  protected image : HTMLImageElement;

  constructor(public imageUrl: string) {
    this.__UID = __UNIQUE_ID__++;
  }

  public GetUniqueID() : string {
    return this.GetTypeName() + btoa(this.__UID + '_' + btoa(this.imageUrl));
  }

  public GetOptimalDimensions() : [number, number] {
    return [0, 0];
  }

  // (Width / Height)
  public GetAspectRatio() : number {
    return 1;
  }

  /* TODO: replace with reflection */
  abstract GetTypeName() : string;

  /* Gives subclasses an opportunity to tweak the dom element */
  abstract GetDomElement() : HTMLElement;
}

export class SeedableRNG {
  private _seed : number;
  constructor() {
    this._seed = Math.random();
  }

  public Seed(n : number) {
    this._seed = n;
  }

  public next(lower?: number, upper?: number) : number {
    this._seed = this._seed * 16807 % 2147483647;
    if (arguments.length === 0) {
      return this._seed / 2147483647;
    }

    if (arguments.length === 1) {
      return (this._seed / 2147483647) * lower;
    }

    if (arguments.length === 2) {
      return (this._seed / 2147483647) * (upper - lower) + lower;
    }
  }
}

class VerticallyExpandingGrid {
  private occupied: number[][];
  private highest_open_row: number;
  private big: boolean;
  private rng: SeedableRNG;

  constructor(private gridColumns: number, private minimumSize: number) {
    this.occupied = []
    this.highest_open_row = 0;
    this.big = true;
    this.rng = new SeedableRNG();
    this.rng.Seed(1236445);
  }

  private rowOpening(row: number): number {
    for(let i=0; i<this.gridColumns; i++) {
      if (!this.isOccupied(i, row)) {
        return i;
      }
    }
    return -1;
  }

  private expandGrid(to: number) {
    while(this.occupied.length-1 < to) {
      this.occupied.push(Array.from({length:this.gridColumns}, ()=>0));
    }
  }

  private openStreakLength(row: number, column: number): number {
    if (column > this.gridColumns) {
      return 0;
    }
    this.expandGrid(row);
    for(let i=column; i<this.gridColumns; i++) {
      if (this.isOccupied(i, row)) {
        return i - column;
      }
    }
    return this.gridColumns - column;
  }

  private canFitWithDimensions(row: number, column: number, width: number, height: number): boolean {
    if (column+width > this.gridColumns) {
      return false;
    }
    this.expandGrid(row+height);
    for(let x=0; x<width; x++) {
      for(let y=0; y<height; y++) {
        if (this.isOccupied(column+x, row+y,)) {
          return false;
        }
      }
    }
    return true;
  }

  private updateHighestOpenRow() {
    while(this.rowOpening(this.highest_open_row) == -1) {
      this.highest_open_row++;
    }
  }

  private isOccupied(x: number, y: number): boolean {
    if (x > this.gridColumns) {
      return false;
    }
    this.expandGrid(y);
    return this.occupied[y][x] != 0;
  }

  public SetTileOccupied(coords: [number/*X*/, number/*Y*/][], by: number) {
    for(let coord of coords) {
      if (coord[0] > this.gridColumns) {
        throw new Error('Out of bounds coordinate ' + coord);
      }
      this.expandGrid(coord[1]);
      this.occupied[coord[1]][coord[0]] = by;
    }

    this.updateHighestOpenRow();
  }

  public setOccupiedRange(rects: [[number, number], [number, number]][]) {
    for (let r of rects) {
      for(let x=r[0][0]; x<r[1][0]; x++) {
        for(let y=r[0][1]; y<r[1][1]; y++) {
          this.SetTileOccupied([[x, y]], -1);
        }
      }
    }
  }

  private placeElement(eid, elements, row, column, update) : GalleryImage[] {
    let _remaining = this.openStreakLength(row, column);
    let _max_size = Math.floor(this.gridColumns / 2.8);
    let _min_size = this.minimumSize;
    if (_min_size * 2 > _remaining) {
      _min_size = _remaining;
    }
    let _pref_size = Math.floor(this.rng.next(_min_size, _max_size));
    if (_remaining - _pref_size < this.minimumSize) {
      _pref_size = _remaining;
    }

    // Try to find a matching element
    for(let idx in elements) {
      let _dims = elements[idx].GetOptimalDimensions();
      if (_dims[0] == _pref_size && this.canFitWithDimensions(row, column, _dims[0], _dims[1])) {
        for(let x=0; x<_dims[0]; x++) {
          for(let y=0; y<_dims[1]; y++) {
            this.SetTileOccupied([[column+x, row+y]], eid);
          }
        }
        update(elements.splice(idx, 1)[0], column, row, _dims[0], _dims[1]);
        return elements;
      }
    }

    // Try to find a dynamically sizable element
    while(_pref_size >= this.minimumSize) {
      for(let idx in elements) {
        let _dims = elements[idx].GetOptimalDimensions();
        let _expected_height = Math.max(1, ~~(_pref_size / elements[idx].GetAspectRatio()));
        if (_dims[0] == 0 && this.canFitWithDimensions(row, column, _pref_size, _expected_height)) {
          for(let x=0; x<_pref_size; x++) {
            for(let y=0; y<_expected_height; y++) {
              this.SetTileOccupied([[column+x, row+y]], eid);
            }
          }
          let _element = elements.splice(idx, 1);
          update(_element[0], column, row, _pref_size, _expected_height);
          return elements;
        }
      }
      _pref_size--;
    }

    this.SetTileOccupied([[column, row]], -1);
    this.updateHighestOpenRow()
    return this.placeElement(eid, elements, 
      this.highest_open_row, this.rowOpening(this.highest_open_row), update);
  }

  public ArrangeElements(elements: GalleryImage[], update: (img: GalleryImage, x, y, w, h)=>void) {
    while(elements.length > 0) {
      let _row_opening = this.rowOpening(this.highest_open_row);
      elements = this.placeElement(elements.length, elements, this.highest_open_row, _row_opening, update);
      this.updateHighestOpenRow();
    }
    return this.occupied.length;
  }

}

export class Gallery {
  private rawImages : GalleryImage[];
  private gridContainer : HTMLElement;
  private generated_rows: number;
  private tileCoords : [number, number][];
  private imageElementMap : Map<string, [GalleryImage, HTMLElement]>;
  private eventHandlers : Map<string, Map<string, (img: GalleryImage)=>void>>;
  private listenerInterceptors : Map<string, (event)=>void>;
  private rects : [[number, number], [number, number]][];

  /* default constructor */
  constructor(private galleryElementId: string, private gridColumns: number, private minimumSize: number) {
    this.imageElementMap = new Map();
    this.eventHandlers = new Map();
    this.listenerInterceptors = new Map();
    this.rawImages = [];
    this.generated_rows = 1;
    this.tileCoords = []
    this.rects = []

    let _element = document.getElementById(galleryElementId);
    if (!_element) {
      throw new Error("Must provide a valid element ID");
    }
    
    _element.style['grid-template-columns'] = "repeat(" + gridColumns + ", 1fr)"
    _element.style['display'] = 'grid'
    this.gridContainer = _element;

    new ResizeObserver(() => {
      this.resetRowHeight(this.generated_rows);
    }).observe(_element);
  }

  private generateUniqueId() : string {
    return '_' + Math.random().toString(36).substr(2, 9);
  }

  private resetRowHeight(rows) {
    let _dim = this.gridContainer.offsetWidth / this.gridColumns;
    this.gridContainer.style['grid-template-rows'] = "repeat(" + rows + ", " + _dim + "px)";
  }

  private repaint() {
    let _g = new VerticallyExpandingGrid(this.gridColumns, this.minimumSize);
    _g.setOccupiedRange(this.rects);
    _g.SetTileOccupied(this.tileCoords, -1);
    let _this_capture = this;
    this.generated_rows = _g.ArrangeElements(this.rawImages.slice(), (E, x, y, w, h)=>{
      let _element = _this_capture.imageElementMap[E.GetUniqueID()][1];
      _element.style['grid-column-start'] = x+1;
      _element.style['grid-column-end'] = w+x+1;
      _element.style['grid-row-start'] = y+1;
      _element.style['grid-row-end'] = h+y+1;
    });
    this.resetRowHeight(this.generated_rows);
  }

  private createNewListenerInterception(event: string) {
    let _this_capture = this;
    if (!this.listenerInterceptors.has(event)) {
      this.listenerInterceptors.set(event, (e) => {
        let _gallery_image : GalleryImage = _this_capture.imageElementMap[e.target.id][0];
        if(_this_capture.eventHandlers.has(event)) {
          for(let id in _this_capture.eventHandlers.get(event)) {
            _this_capture.eventHandlers.get(event).get(id)(_gallery_image);
          }
        }
      });
      let _event = this.listenerInterceptors.get(event);
      for(let id in this.imageElementMap) {
        this.imageElementMap[id][1].addEventListener(event, _event);
      }
    }
  }

  private applyListeners(element: HTMLElement) {
    for(let event in this.listenerInterceptors) {
      element.addEventListener(event, this.listenerInterceptors[event]);
    }
  }

  public SetExcludedTiles(tileCoords: [number, number][]) {
    this.tileCoords = tileCoords;
    this.repaint();
  }

  public SetExcludedRange(rect: [[number, number], [number, number]]) {
    this.rects.push(rect);
  }

  public Render(images: GalleryImage[]) {
    this.rawImages = this.rawImages.concat(images);
    for(let img of images) {
      let _dom_element = img.GetDomElement();
      this.applyListeners(_dom_element);
      this.gridContainer.appendChild(_dom_element);
      this.imageElementMap[img.GetUniqueID()] = [img, _dom_element];
    }
    this.repaint();
  }

  public OnEvent(event: string, callback: (GalleryImage) => void) : string {
    if (!(event in this.eventHandlers)) {
      this.createNewListenerInterception(event);
      this.eventHandlers[event] = new Map();
    }
    let _result_key = this.generateUniqueId();
    this.eventHandlers[event][_result_key] = callback;
    return _result_key;
  }

  public StopListening(event: string, listener_id: string) {
    if (this.eventHandlers.has(event)) {
      if (this.eventHandlers[event].has(listener_id)) {
        this.eventHandlers[event].delete(listener_id);
      }
    }
  }
}